package cantabularmetadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

// QueryDefaultClassification is the graphQL query to determine which of the provided variables
// is the default categorisation
const QueryDefaultClassification = `
query ($dataset: String!, $variables: [String!]!) {
	dataset(name: $dataset) {
		vars(names: $variables) {
		name
		meta {
				Default_Classification_Flag
			}
		} 
	}
}`

// QueryData holds all the possible required variables to encode any of the graphql queries defined in this file.
type QueryData struct {
	Dataset   string
	Variables []string
}

// Encode the provided graphQL query with the data in QueryData
// returns a byte buffer with the encoded query, along with any encoding error that might happen
func (data *QueryData) Encode(query string) (bytes.Buffer, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)

	vars := map[string]interface{}{
		"dataset":   data.Dataset,
		"variables": data.Variables,
	}

	if err := enc.Encode(map[string]interface{}{
		"query":     query,
		"variables": vars,
	}); err != nil {
		return b, fmt.Errorf("failed to encode GraphQL query: %w", err)
	}

	return b, nil
}

// queryUnmarshal uses postQuery to perform a graphQL query and then un-marshals the response body to the provided value pointer v
// This method handles the response body closing.
func (c *Client) queryUnmarshal(ctx context.Context, graphQLQuery string, data QueryData, v interface{}) error {
	url := fmt.Sprintf("%s/graphql", c.host)

	logData := log.Data{
		"url":        url,
		"query":      graphQLQuery,
		"query_data": data,
	}

	res, err := c.postQuery(ctx, graphQLQuery, data)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to post query: %s", err),
			http.StatusInternalServerError,
			logData,
		)
	}
	defer closeResponseBody(ctx, res)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			c.StatusCode(err),
			logData,
		)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return nil
}

// postQuery performs a query against the Cantabular Extended API
// using the /graphql endpoint and the http client directly
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) postQuery(ctx context.Context, graphQLQuery string, data QueryData) (*http.Response, error) {
	url := fmt.Sprintf("%s/graphql", c.host)

	logData := log.Data{
		"url": url,
	}

	b, err := data.Encode(graphQLQuery)
	logData["query"] = b.String()
	if err != nil {
		return nil, dperrors.New(err, http.StatusInternalServerError, logData)
	}

	// Do a POST call to graphQL endpoint
	res, err := c.httpPost(ctx, url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			c.StatusCode(err),
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
