package cantabular

import (
	"context"

	"github.com/shurcooL/graphql"
)

type PopulationTypeQueryDataset struct {
	Name graphql.String
}

type PopulationTypeQuery struct {
	Datasets []PopulationTypeQueryDataset
}

func (c *Client) GetPopulationTypes(ctx context.Context) ([]string, error) {
	var query PopulationTypeQuery
	if err := c.gqlClient.Query(ctx, &query, nil); err != nil {
		return nil, err
	}
	names := make([]string, len(query.Datasets))
	for i, dataset := range query.Datasets {
		names[i] = string(dataset.Name)
	}
	return names, nil
}
