package cantabular

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

// GetDimensions performs a graphQL query to obtain all the dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetDimensions(ctx context.Context, dataset string) (*GetDimensionsResponse, error) {
	resp := &struct {
		Data   GetDimensionsResponse `json:"data"`
		Errors []gql.Error           `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryDimensions, QueryData{Dataset: dataset}, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetGeographyDimensions performs a graphQL query to obtain the geography dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetGeographyDimensions(ctx context.Context, dataset string) (*GetGeographyDimensionsResponse, error) {
	resp := &struct {
		Data   GetGeographyDimensionsResponse `json:"data"`
		Errors []gql.Error                    `json:"errors,omitempty"`
	}{}

	if err := c.queryUnmarshal(ctx, QueryGeographyDimensions, QueryData{Dataset: dataset}, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetDimensionsByName performs a graphQL query to obtain only the dimensions that match the provided dimension names for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetDimensionsByName(ctx context.Context, req GetDimensionsByNameRequest) (*GetDimensionsResponse, error) {
	resp := &struct {
		Data   GetDimensionsResponse `json:"data"`
		Errors []gql.Error           `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.DimensionNames,
	}

	if err := c.queryUnmarshal(ctx, QueryDimensionsByName, data, resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// GetDimensionOptions performs a graphQL query to obtain the requested dimension options.
// It returns a Table with a list of Cantabular dimensions, where 'Variable' is the dimension and 'Categories' are the options
// The whole response is loaded to memory.
func (c *Client) GetDimensionOptions(ctx context.Context, req GetDimensionOptionsRequest) (*GetDimensionOptionsResponse, error) {
	resp := &struct {
		Data   GetDimensionOptionsResponse `json:"data"`
		Errors []gql.Error                 `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.DimensionNames,
	}

	if err := c.queryUnmarshal(ctx, QueryDimensionOptions, QueryData(data), resp); err != nil {
		return nil, err
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			http.StatusOK,
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}
