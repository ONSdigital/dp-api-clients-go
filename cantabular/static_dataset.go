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
	"github.com/ONSdigital/dp-api-clients-go/v2/jsonstream"
	"github.com/ONSdigital/log.go/v2/log"
)

// // StaticDataset represents the 'dataset' field from a GraphQL static dataset
// // query response
// type StaticDataset struct {
// 	Table Table `json:"table" graphql:"table(variables: $variables)"`
// }

// StaticDatasetQuery performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint
// func (c *Client) StaticDatasetQuery(ctx context.Context, req StaticDatasetQueryRequest) (*StaticDatasetQuery, error) {
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

	res, err := http.Post(url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	if res.StatusCode != http.StatusOK {
		return nil, c.errorResponse(url, res)
	}

	// bb, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println(string(bb))

	return res.Body, nil

	/*
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
	*/
}

// GraphqlJSONToCSV converts a JSON response in r to CSV on w and panics on error
func GraphqlJSONToCSV(r io.Reader, w io.Writer) error {
	dec := jsonstream.New(r)
	if !dec.StartObjectComposite() {
		return errors.New("no json object found in response")
	}
	for dec.More() {
		switch field := dec.DecodeName(); field {
		case "data":
			if dec.StartObjectComposite() {
				decodeDataFields(dec, w)
				dec.EndComposite()
			}
		case "errors":
			if err := decodeErrors(dec); err != nil {
				return err
			}
		}
	}
	dec.EndComposite()
	return nil
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// If no table cell values are present then no output is written.
func decodeTableFields(dec jsonstream.Decoder, w io.Writer) {
	var dims Dimensions
	for dec.More() {
		switch field := dec.DecodeName(); field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				panic(err)
			}
		case "error":
			if errMsg := dec.DecodeString(); errMsg != nil {
				panic(fmt.Sprintf("Table blocked: %s", *errMsg))
			}
		case "values":
			if dims == nil {
				panic("values received before dimensions")
			}
			if dec.StartArrayComposite() {
				decodeValues(dec, dims, w)
				dec.EndComposite()
			}
		}
	}
}
