package cantabular

import (
	"context"

	"github.com/pkg/errors"

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

	data := QueryData{
		Dataset: dataset,
	}

	if err := c.queryUnmarshal(ctx, QueryDimensions, data, resp); err != nil {
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

// GetGeographyDimensions performs a graphQL query to obtain the geography dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetGeographyDimensions(ctx context.Context, req GetGeographyDimensionsRequest) (*GetGeographyDimensionsResponse, error) {
	resp := struct {
		Data   GetGeographyDimensionsResponse `json:"data"`
		Errors []gql.Error                    `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:          req.Dataset,
		PaginationParams: req.PaginationParams,
	}

	if err := c.queryUnmarshal(ctx, QueryGeographyDimensions, data, &resp); err != nil {
		return nil, err
	}

	resp.Data.PaginationResponse = PaginationResponse{
		Count:            len(resp.Data.Dataset.RuleBase.IsSourceOf.Edges),
		TotalCount:       resp.Data.Dataset.RuleBase.IsSourceOf.TotalCount,
		PaginationParams: req.PaginationParams,
	}

	if len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			resp.Errors[0].StatusCode(),
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
			resp.Errors[0].StatusCode(),
			log.Data{"errors": resp.Errors},
		)
	}

	return &resp.Data, nil
}

// SearchDimensionsRequest performs a graphQL query to obtain the dimensions that match the provided text in the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) SearchDimensions(ctx context.Context, req SearchDimensionsRequest) (*GetDimensionsResponse, error) {
	resp := &struct {
		Data   GetDimensionsResponse `json:"data"`
		Errors []gql.Error           `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset: req.Dataset,
		Text:    req.Text,
	}

	if err := c.queryUnmarshal(ctx, QueryDimensionsSearch, data, resp); err != nil {
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
		Filters:   req.Filters,
	}

	if err := c.queryUnmarshal(ctx, QueryDimensionOptions, data, resp); err != nil {
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

// GetAreas performs a graphQL query to retrieve the areas (categories) for a given area type. If the category
// is left empty, then all categories are returned. Results can also be filtered by area by passing a variable name.
func (c *Client) GetAreas(ctx context.Context, req GetAreasRequest) (*GetAreasResponse, error) {
	resp := &struct {
		Data   GetAreasResponse `json:"data"`
		Errors []gql.Error      `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:  req.Dataset,
		Text:     req.Variable,
		Category: req.Category,
	}

	if err := c.queryUnmarshal(ctx, QueryAreas, data, resp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal query")
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
