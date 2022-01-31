package cantabular

import (
	"context"

	"github.com/shurcooL/graphql"
)

type ListDatasetsItem struct {
	Name graphql.String
}

type ListDatasetsQuery struct {
	Datasets []ListDatasetsItem
}

func (c *Client) ListDatasets(ctx context.Context) ([]string, error) {
	var query ListDatasetsQuery
	if err := c.gqlClient.Query(ctx, &query, nil); err != nil {
		return nil, err
	}
	names := make([]string, len(query.Datasets))
	for i, dataset := range query.Datasets {
		names[i] = string(dataset.Name)
	}
	return names, nil
}
