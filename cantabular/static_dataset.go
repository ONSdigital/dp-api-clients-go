package cantabular

import (
	"context"
)

// GQLDataset represents the 'dataset' field from a GraphQL static dataset
// query response
type StaticDataset struct{
	Table Table `json:"table"`
}

// StaticDatasetQuery performs a query for a static dataset against the
// Cantabular Extended API using the /graphql endpoint
func (c *Client) StaticDatasetQuery(ctx context.Context, req StaticDatasetQueryRequest) (*StaticDatasetQueryResponse, error){
	return nil, nil
}