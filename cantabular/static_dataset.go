package cantabular

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/shurcooL/graphql"
)

// StaticDataset represents the 'dataset' field from a GraphQL static dataset
// query response
type StaticDataset struct {
	Table Table `json:"table" graphql:"table(variables: $variables)"`
}

// StaticDatasetQuery performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint
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
