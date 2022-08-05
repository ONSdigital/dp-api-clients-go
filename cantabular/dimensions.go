package cantabular

import (
	"context"

	"github.com/ONSdigital/dp-api-clients-go/v2/batch"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/pkg/errors"
)

// GetGeographyBatchProcessor is the type corresponding to a batch processing function for Geography dimensions
type GetGeographyBatchProcessor func(response *GetGeographyDimensionsResponse) (abort bool, err error)

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
		Count:            len(resp.Data.Dataset.Variables.Edges),
		TotalCount:       resp.Data.Dataset.Variables.TotalCount,
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
		PaginationParams: req.PaginationParams,
		Dataset:          req.Dataset,
		Text:             req.Variable,
		Category:         req.Category,
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

// GetParents returns a list of variables that map to the provided variable
func (c *Client) GetParents(ctx context.Context, req GetParentsRequest) (*GetParentsResponse, error) {
	resp := &struct {
		Data   GetParentsResponse `json:"data"`
		Errors []gql.Error        `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:   req.Dataset,
		Variables: []string{req.Variable},
	}

	if err := c.queryUnmarshal(ctx, QueryParents, data, resp); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal query")
	}

	if resp != nil && len(resp.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			resp.Errors[0].StatusCode(),
			log.Data{
				"request": req,
				"errors":  resp.Errors,
			},
		)
	}

	// should be impossible but to avoid panic
	if len(resp.Data.Dataset.Variables.Edges) < 1 {
		return nil, errors.New("invalid response from graphQL")
	}

	// last item is guaranteed to be provided variable, only return parents
	edges := resp.Data.Dataset.Variables.Edges[0].Node.IsSourceOf.Edges
	resp.Data.Dataset.Variables.Edges[0].Node.IsSourceOf.Edges = edges[:len(edges)-1]

	return &resp.Data, nil
}

// GetGeographyDimensionsInBatches performs a graphQL query to obtain all the geography dimensions for the provided cantabular dataset.
// The whole response is loaded to memory.
func (c *Client) GetGeographyDimensionsInBatches(ctx context.Context, datasetID string, batchSize, maxWorkers int) (*gql.Dataset, error) {
	// reference GetInstanceDimensionsInBatches
	var dataset *gql.Dataset
	var processBatch GetGeographyBatchProcessor = func(b *GetGeographyDimensionsResponse) (bool, error) {
		if dataset == nil {
			dataset = &gql.Dataset{
				Variables: gql.Variables{
					TotalCount: b.Dataset.Variables.TotalCount,
					Edges:      make([]gql.Edge, b.Dataset.Variables.TotalCount),
				},
			}
		}

		for i := range b.Dataset.Variables.Edges {
			dataset.Variables.Edges[i+b.PaginationResponse.Offset] = b.Dataset.Variables.Edges[i]
		}
		return false, nil
	}

	// call GetGeographyBatchProcess in batches and aggregate the responses
	err := c.GetGeographyBatchProcess(ctx, datasetID, processBatch, batchSize, maxWorkers)
	if err != nil {
		return nil, errors.Wrap(err, "GetGeographyBatchProcess failed")
	}

	return dataset, nil
}

// GetGeographyBatchProcess gets the geography dimensions from the API in batches, calling the provided function for each batch.
func (c *Client) GetGeographyBatchProcess(ctx context.Context, datasetID string, processBatch GetGeographyBatchProcessor, batchSize, maxWorkers int) error {
	// for each batch, obtain the dimensions starting at the provided offset, with a batch size limit
	batchGetter := func(offset int) (interface{}, int, string, error) {
		req := GetGeographyDimensionsRequest{
			PaginationParams: PaginationParams{
				Offset: offset,
				Limit:  batchSize,
			},
			Dataset: datasetID,
		}

		b, err := c.GetGeographyDimensions(ctx, req)
		if err != nil {
			return nil, 0, "", errors.Wrapf(err, "GetGeographyDimensions failed for offset: %d", offset)
		}

		return b, b.TotalCount, "", nil
	}

	// cast and process the batch according to the provided method
	batchProcessor := func(b interface{}, _ string) (bool, error) {
		v, ok := b.(*GetGeographyDimensionsResponse)
		if !ok {
			return true, errors.New("wrong type")
		}
		return processBatch(v)
	}

	return batch.ProcessInConcurrentBatches(batchGetter, batchProcessor, batchSize, maxWorkers)
}
