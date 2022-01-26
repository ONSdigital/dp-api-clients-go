package cantabular

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
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

// GetDimensions performs a graphQL query to obtain all the dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetDimensions(ctx context.Context, dataset string) (*GetDimensionsResponse, error) {
	resp := &struct {
		Data   GetDimensionsResponse `json:"data"`
		Errors []gql.Error           `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryDimensions, QueryData{Dataset: dataset}, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetGeographyDimensions performs a graphQL query to obtain the geography dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetGeographyDimensions(ctx context.Context, dataset string) (*GetGeographyDimensionsResponse, error) {
	resp := &struct {
		Data   GetGeographyDimensionsResponse `json:"data"`
		Errors []gql.Error                    `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryGeographyDimensions, QueryData{Dataset: dataset}, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetDimensions performs a graphQL query to obtain the dimensions for the provided cantabular dataset.
// It returns a RuleBaseResponse, containing nested edges and nodes according to the query structure
// The whole response is loaded to memory.
func (c *Client) GetDimensionsByName(ctx context.Context, req GetDimensionsByNameRequest) (*GetDimensionsResponse, error) {
	resp := &struct {
		Data   GetDimensionsResponse `json:"data"`
		Errors []gql.Error           `json:"errors,omitempty"`
	}{}
	if err := c.queryUnmarshal(ctx, QueryDimensionsByName, QueryData(req), resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetDimensionOptions performs a graphQL query to obtain the requested dimension options.
// It returns a Table with a list of Cantabular dimensions, where 'Variable' is the dimension and 'Categories' are the options
// The whole response is loaded to memory.
func (c *Client) GetDimensionOptions(ctx context.Context, req GetDimensionOptionsRequest) (*GetDimensionOptionsResponse, error) {
	resp := &struct {
		Data   GetDimensionOptionsResponse `json:"data"`
		Errors []gql.Error                 `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryDimensionOptions, QueryData(req), resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// StaticDatasetQueryStreamCSV performs a StaticDatasetQuery call
// and then starts 2 go-routines to transform the response body into a CSV stream and
// consume the transformed output with the provided Consumer concurrently.
// The number of CSV rows, including the header, is returned along with any error during the process.
// Use this method if large query responses are expected.
func (c *Client) StaticDatasetQueryStreamCSV(ctx context.Context, req StaticDatasetQueryRequest, consume Consumer) (int32, error) {
	res, err := c.postQuery(ctx, QueryStaticDataset, QueryData(req))
	if err != nil {
		return 0, err
	}
	var rowCount int32

	transform := func(ctx context.Context, body io.Reader, pipeWriter io.Writer) error {
		if rowCount, err = GraphQLJSONToCSV(ctx, body, pipeWriter); err != nil {
			return err
		}
		return nil
	}
	return rowCount, stream.Stream(ctx, res.Body, transform, consume)
}

// queryUnmarshal uses postQuery to perform a graphQL query and then un-marshals the response body to the provided value pointer v
// This method handles the response body closing.
func (c *Client) queryUnmarshal(ctx context.Context, graphQLQuery string, data QueryData, v interface{}) error {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	res, err := c.postQuery(ctx, graphQLQuery, data)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, res)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			res.StatusCode,
			log.Data{
				"url": url,
			},
		)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{
				"url":           url,
				"response_body": string(b),
			},
		)
	}

	return nil
}

// postQuery performs a query against the Cantabular Extended API
// using the /graphql endpoint and the http client directly
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) postQuery(ctx context.Context, graphQLQuery string, data QueryData) (*http.Response, error) {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url":        url,
		"query_data": data,
	}

	b, err := data.Encode(graphQLQuery)
	if err != nil {
		return nil, dperrors.New(err, http.StatusInternalServerError, logData)
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

	return res, nil
}
