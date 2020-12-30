package filter

import (
	"context"
	"errors"

	"github.com/ONSdigital/dp-api-clients-go/common"
)

// BatchProcessor is the type corresponding to a batch processing function for filter DimensionOptions
type BatchProcessor func(DimensionOptions) (abort bool, err error)

// GetDimensionOptionsInBatches retrieves a list of the dimension options in concurrent batches and accumulates the results
func (c *Client) GetDimensionOptionsInBatches(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, batchSize, maxWorkers int) (opts DimensionOptions, err error) {

	// function to aggregate items. The first received batch will initialise the structure with items with a fixed size
	// and items are aggregated in order even if batches are received in the wrong order (e.g. due to concurrent calls)
	var processBatch BatchProcessor = func(batch DimensionOptions) (abort bool, err error) {
		if len(opts.Items) == 0 {
			opts.TotalCount = batch.TotalCount
			opts.Items = make([]DimensionOption, batch.TotalCount)
			opts.Count = batch.TotalCount
		}
		for i := 0; i < len(batch.Items); i++ {
			opts.Items[i+batch.Offset] = batch.Items[i]
		}
		return false, nil
	}

	// call filter API GetOptions in batches and aggregate the responses
	if err := c.BatchProcessDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, processBatch, batchSize, maxWorkers); err != nil {
		return DimensionOptions{}, err
	}
	return opts, nil
}

// BatchProcessDimensionOptions gets the filter options for a dimension from filter API in batches, and calls the provided function for each batch.
func (c *Client) BatchProcessDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, processBatch BatchProcessor, batchSize, maxWorkers int) (err error) {

	// for each batch, obtain the dimensions starting at the provided offset, with a batch size limit
	batchGetter := func(offset int) (interface{}, int, error) {
		batch, err := c.GetDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, QueryParams{Offset: offset, Limit: batchSize})
		return batch, batch.TotalCount, err
	}

	// cast and process the batch according to the provided method
	batchProcessor := func(batch interface{}) (abort bool, err error) {
		v, ok := batch.(DimensionOptions)
		if !ok {
			return true, errors.New("wrong type")
		}
		return processBatch(v)
	}

	return common.ProcessInConcurrentBatches(batchGetter, batchProcessor, batchSize, maxWorkers)
}
