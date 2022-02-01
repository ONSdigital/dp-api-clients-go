package cantabular

import (
	"context"
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

// StaticDatasetDimensionOptions represents the 'dataset' field from a GraphQL static dataset
// query response, containing a DimensionsTable, without values
type StaticDatasetDimensionOptions struct {
	Table DimensionsTable `json:"table"`
}

// DimensionsTable represents the 'table' field from the GraphQL dataset response,
// which contains only dimensions and error fields
type DimensionsTable struct {
	Dimensions []Dimension `json:"dimensions"`
	Error      string      `json:"error,omitempty" `
}

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
	res, err := c.postQuery(ctx, QueryStaticDataset, QueryData(req))
	if err != nil {
		closeResponseBody(ctx, res) // close response body, as it is not passed to the Stream func
		return 0, err
	}
	var rowCount int32

	// transform will be executed by Stream when processing the data into 'csv' format.
	transform := func(ctx context.Context, body io.Reader, pipeWriter io.Writer) error {
		if rowCount, err = GraphQLJSONToCSV(ctx, body, pipeWriter); err != nil {
			return err
		}
		return nil
	}

	// Stream is responsible for closing the response body
	return rowCount, stream.Stream(ctx, res.Body, transform, consume)
}
