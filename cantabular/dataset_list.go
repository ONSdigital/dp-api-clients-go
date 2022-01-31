package cantabular

import (
	"context"

	"github.com/shurcooL/graphql"
)

type DatasetListItem struct {
	Name graphql.String
}

type DatasetListQuery struct {
	Datasets []DatasetListItem
}

func (c *Client) ListDatasets(ctx context.Context) ([]string, error) {
	var query DatasetListQuery
	if err := c.gqlClient.Query(ctx, &query, nil); err != nil {
		return nil, err
	}
	names := make([]string, len(query.Datasets))
	for i, dataset := range query.Datasets {
		names[i] = string(dataset.Name)
	}
	return names, nil
}
