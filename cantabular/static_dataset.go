package cantabular

import (
	"context"
	"fmt"
	"io"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-api-clients-go/v2/stream"
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
	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	var q struct {
		Data struct{ *StaticDatasetQuery } `json:"data"`
	}
	if err := c.queryUnmarshal(ctx, QueryStaticDataset, QueryData{Dataset: req.Dataset, Variables: req.Variables, Filters: req.Filters}, &q); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return q.Data.StaticDatasetQuery, nil
}

// StaticDatasetQueryStreamCSV performs a StaticDatasetQuery call
// and then starts 2 go-routines to transform the response body into a CSV stream and
// consume the transformed output with the provided Consumer concurrently.
// The number of CSV rows, including the header, is returned along with any error during the process.
// Use this method if large query responses are expected.
func (c *Client) StaticDatasetQueryStreamCSV(ctx context.Context, req StaticDatasetQueryRequest, consume Consumer) (int32, error) {
	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.Variables,
		Filters:   req.Filters,
	}

	res, err := c.postQuery(ctx, QueryStaticDataset, data)
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
