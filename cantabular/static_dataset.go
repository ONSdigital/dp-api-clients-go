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
	isStartObj, err := dec.StartObjectComposite()
	if err != nil {
		return fmt.Errorf("error decoding start of json object: %w", err)
	}

	if !isStartObj {
		return errors.New("no json object found in response")
	}
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "data":
			isStartObj, err := dec.StartObjectComposite()
			if err != nil {
				return fmt.Errorf("error decoding start of json object for 'data': %w", err)
			}
			if isStartObj {
				if err := decodeDataFields(dec, w); err != nil {
					return fmt.Errorf("error decoding data fields: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return fmt.Errorf("error decoding end of json object for 'data': %w", err)
				}
			}
		case "errors":
			if err := decodeErrors(dec); err != nil {
				return err
			}
		}
	}
	if err := dec.EndComposite(); err != nil {
		return fmt.Errorf("error decoding end of json object: %w", err)
	}
	return nil
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// If no table cell values are present then no output is written.
func decodeTableFields(dec jsonstream.Decoder, w io.Writer) error {
	var dims Dimensions
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				return fmt.Errorf("error decoding dimensions: %w", err)
			}
		case "error":
			errMsg, err := dec.DecodeString()
			if err != nil {
				return fmt.Errorf("error decoding error message: %w", err)
			}
			if errMsg != nil {
				return fmt.Errorf("table blocked: %s", *errMsg)
			}
		case "values":
			if dims == nil {
				return errors.New("values received before dimensions")
			}
			isStartArray, err := dec.StartArrayComposite()
			if err != nil {
				return fmt.Errorf("error decoding start of json array for 'values': %w", err)
			}
			if isStartArray {
				if err := decodeValues(dec, dims, w); err != nil {
					return fmt.Errorf("error decoding values: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return fmt.Errorf("error decoding end of json array for 'values': %w", err)
				}
			}
		}
	}
	return nil
}
