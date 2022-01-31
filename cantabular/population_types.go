package cantabular

import (
	"context"

	"github.com/shurcooL/graphql"
)

type BlobQueryDataset struct {
	Name graphql.String
}

type BlobQuery struct {
	Datasets []BlobQueryDataset
}

func (c *Client) GetBlobs(ctx context.Context) ([]string, error) {
	var query BlobQuery
	if err := c.gqlClient.Query(ctx, &query, nil); err != nil {
		return nil, err
	}
	names := make([]string, len(query.Datasets))
	for i, dataset := range query.Datasets {
		names[i] = string(dataset.Name)
	}
	return names, nil
}
