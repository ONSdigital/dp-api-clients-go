package cantabular

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/stream"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/pkg/errors"
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

type StaticDatasetQueryTypeResponse struct {
	Dataset gql.Dataset `json:"dataset"`
}

// StaticDatasetType will return the type of dataset
func (c *Client) StaticDatasetType(ctx context.Context, datasetName string) (*gql.Dataset, error) {
	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": datasetName,
	}

	var q struct {
		Data   StaticDatasetQueryTypeResponse `json:"data"`
		Errors []gql.Error                    `json:"errors"`
	}

	qd := QueryData{
		Dataset: datasetName,
	}

	if err := c.queryUnmarshal(ctx, QueryStaticDatasetType, qd, &q); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	if len(q.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			q.Errors[0].StatusCode(),
			log.Data{"errors": q.Errors},
		)
	}

	return &q.Data.Dataset, nil
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
		Data   StaticDatasetQuery `json:"data"`
		Errors []gql.Error        `json:"errors"`
	}

	qd := QueryData{
		Dataset:   req.Dataset,
		Variables: req.Variables,
		Filters:   req.Filters,
	}

	if err := c.queryUnmarshal(ctx, QueryStaticDataset, qd, &q); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	if len(q.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			q.Errors[0].StatusCode(),
			log.Data{"errors": q.Errors},
		)
	}

	if len(q.Data.Dataset.Table.Error) != 0 {
		return nil, dperrors.New(
			errors.New(c.parseTableError(q.Data.Dataset.Table.Error)),
			http.StatusBadRequest,
			logData,
		)
	}

	return &q.Data, nil
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

// Checks the number of observations returned from a cantabular query
func (c *Client) CheckQueryCount(ctx context.Context, req StaticDatasetQueryRequest) (int, error) {
	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.Variables,
		Filters:   req.Filters,
	}

	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	var q struct {
		Data   StaticDatasetQuery `json:"data"`
		Errors []gql.Error        `json:"errors"`
	}

	if err := c.queryUnmarshal(ctx, QueryStaticDataset, data, &q); err != nil {
		return 0, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	rowCount := len(q.Data.Dataset.Table.Values)

	if len(q.Data.Dataset.Table.Error) > 0 || q.Data.Dataset.Table.Values == nil {
		return 0, dperrors.New(
			errors.New(c.parseTableError(q.Data.Dataset.Table.Error)),
			http.StatusBadRequest,
			logData,
		)
	}

	return rowCount, nil
}

// StaticDatasetQueryStreamJson performs a StaticDatasetQuery call
// and then starts 2 go-routines to transform the response body into a Json stream and
// consume the transformed output with the provided Consumer concurrently.
// Returns a json formatted response
// Use this method if large query responses are expected.
func (c *Client) StaticDatasetQueryStreamJson(ctx context.Context, req StaticDatasetQueryRequest, consume Consumer) (GetObservationsResponse, error) {
	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.Variables,
		Filters:   req.Filters,
	}

	res, err := c.postQuery(ctx, QueryStaticDataset, data)
	if err != nil {
		closeResponseBody(ctx, res) // close response body, as it is not passed to the Stream func
		return GetObservationsResponse{}, err
	}

	defer func() { _ = res.Body.Close() }()

	var getObservationsResponse GetObservationsResponse

	// transform will be executed by Stream when processing the data into 'json' format.
	transform := func(ctx context.Context, body io.Reader, pipeWriter io.Writer) error {
		if getObservationsResponse, err = GraphQLJSONToJson(ctx, body, pipeWriter); err != nil {
			return err
		}
		return nil
	}

	// Stream is responsible for closing the response body
	return getObservationsResponse, stream.Stream(ctx, res.Body, transform, consume)

}
