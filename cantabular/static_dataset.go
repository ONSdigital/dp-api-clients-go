package cantabular

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/stream"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/shurcooL/graphql"
)

// Consumer is a stream func to read from a reader
type Consumer = stream.Consumer

// StaticDataset represents the 'dataset' field from a GraphQL static dataset
// query response
type StaticDataset struct {
	Table Table `json:"table" graphql:"table(variables: $variables)"`
}

// QueryStaticDataset represents a full graphQL query to be encoded into the body of a plain http post request
// for a static dataset query
const QueryStaticDataset = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				count
				variable { name label }
				categories { code label } }
			values
			error
		  }
	 }
}`

// StaticDatasetQuery performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint and returns a StaticDatasetQuery,
// loading the whole response to memory.
// Use this method only if large query responses are NOT expected
func (c *Client) StaticDatasetQuery(ctx context.Context, req StaticDatasetQueryRequest) (*StaticDatasetQuery, error) {
	if c.gqlClient == nil {
		return nil, dperrors.New(
			errors.New("cantabular Extended API Client not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	vars := map[string]interface{}{
		// GraphQL package requires self defined scalar types for variables
		// and arguments
		"name": graphql.String(req.Dataset),
	}

	gvars := make([]graphql.String, 0)

	for _, v := range req.Variables {
		gvars = append(gvars, graphql.String(v))
	}

	vars["variables"] = gvars

	var q StaticDatasetQuery
	if err := c.gqlClient.Query(ctx, &q, vars); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	if err := q.Dataset.Table.Error; err != "" {
		return nil, dperrors.New(
			fmt.Errorf("GraphQL error: %s", err),
			http.StatusBadRequest,
			logData,
		)
	}

	return &q, nil
}

// StaticDatasetQueryStreamCSV performs a StaticDatasetQuery call
// and then starts 2 go-routines to transform the response body into a CSV stream and
// consume the transformed output with the provided Consumer concurrently.
// The number of CSV rows, including the header, is returned along with any error during the process.
// Use this method if large query responses are expected.
func (c *Client) StaticDatasetQueryStreamCSV(ctx context.Context, req StaticDatasetQueryRequest, consume Consumer) (int32, error) {
	responseBody, err := c.staticDatasetQueryLowLevel(ctx, req)
	if err != nil {
		return 0, err
	}
	var rowCount int32

	transform := func(ctx context.Context, r io.Reader, w io.Writer) error {
		if rowCount, err = GraphQLJSONToCSV(r, w); err != nil {
			return err
		}
		return nil
	}
	return rowCount, stream.Stream(ctx, responseBody, transform, consume)
}

// staticDatasetQueryLowLevel performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint and the http client directly
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) staticDatasetQueryLowLevel(ctx context.Context, req StaticDatasetQueryRequest) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url":     url,
		"request": req,
	}

	// Encoder the graphQL query with the provided dataset and variables
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": QueryStaticDataset,
		"variables": map[string]interface{}{
			"dataset":   req.Dataset,
			"variables": req.Variables,
		},
	}); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to encode GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	// Do a POST call to graphQL endpoint
	res, err := c.httpPost(ctx, url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	// Check status code and return error
	if res.StatusCode != http.StatusOK {
		closeResponseBody(ctx, res)
		return nil, c.errorResponse(url, res)
	}

	return res.Body, nil
}
