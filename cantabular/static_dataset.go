package cantabular

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

// // StaticDataset represents the 'dataset' field from a GraphQL static dataset
// // query response
// type StaticDataset struct {
// 	Table Table `json:"table" graphql:"table(variables: $variables)"`
// }

// StaticDatasetQuery performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) StaticDatasetQuery(ctx context.Context, req StaticDatasetQueryRequest) (io.ReadCloser, error) {
	if c.gqlClient == nil {
		return nil, dperrors.New(
			errors.New("cantabular Extended API Client not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url":     url,
		"request": req,
	}

	const graphQLQuery = `
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

	// Create a json stream encoder with the query and provided dataset and variables
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": graphQLQuery,
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

	// Do a POST call to graphQL endpoint where body will be written to the same buffer used by the json stream encoder
	res, err := http.Post(url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	// Check status code and return error
	if res.StatusCode != http.StatusOK {
		closeResponseBody(ctx, res.Body)
		return nil, c.errorResponse(url, res)
	}

	return res.Body, nil
}

type Transformer = func(ctx context.Context, r io.Reader, w io.Writer) error
type Consumer = func(ctx context.Context, r io.Reader) error

func (c *Client) StaticDatasetQueryStream(ctx context.Context, req StaticDatasetQueryRequest, transform Transformer, consume Consumer) error {
	responseBody, err := c.StaticDatasetQuery(ctx, req)
	if err != nil {
		return err
	}

	r, w := io.Pipe()
	wg := &sync.WaitGroup{}

	// Start go-routine to read response body, transform it 'on-the-fly' and write it to the pipe writer
	wg.Add(1)
	go func() {
		defer func() {
			closeResponseBody(ctx, responseBody)
			w.Close()
			wg.Done()
		}()
		err = transform(ctx, responseBody, w)
	}()

	// Start go-routine to read pipe reader (transformed) and call the consumer func
	wg.Add(1)
	go func() {
		defer func() {
			r.Close()
			wg.Done()
		}()
		err = consume(ctx, r)
	}()

	wg.Wait()
	return err
}

func (c *Client) StaticDatasetQueryStreamCSV(ctx context.Context, req StaticDatasetQueryRequest, consume Consumer) error {
	transform := func(ctx context.Context, r io.Reader, w io.Writer) error {
		return GraphqlJSONToCSV(r, w)
	}
	return c.StaticDatasetQueryStream(ctx, req, transform, consume)
}
