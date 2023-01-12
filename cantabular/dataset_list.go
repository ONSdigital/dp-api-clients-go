package cantabular

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

func (c *Client) ListDatasets(ctx context.Context) (*ListDatasetsResponse, error) {
	resp := &struct {
		Data   ListDatasetsResponse `json:"data"`
		Errors []gql.Error          `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryListDatasets, QueryData{}, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			resp.Errors[0].StatusCode(),
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}
