package filter

import (
	"context"
	"sync"
)

// BatchProcessor is the type corresponding to a batch processing function for filter DimensionOptions
type BatchProcessor func(DimensionOptions) (abort bool, err error)

// GetDimensionOptionsInBatches retrieves a list of the dimension options in batches and accumulates the results
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
	wg := sync.WaitGroup{}
	chWait := make(chan struct{})
	chErr := make(chan error)
	chAbort := make(chan struct{})
	chSemaphore := make(chan struct{}, maxWorkers)

	lockResult := sync.Mutex{}

	// worker add delta to workgroup and acquire semaphore
	acquire := func() {
		wg.Add(1)
		chSemaphore <- struct{}{}
	}

	// worker release semaphore and workgroup delta
	release := func() {
		<-chSemaphore
		wg.Done()
	}

	// initial state of control variables
	totalCount := 1
	first := true
	abort := false

	// func executed in each go-routine to process the batch, aggregate results, and send errors to the error channel
	doProcessBatch := func(offset int) {
		defer release()

		// Abort if needed
		if abort {
			return
		}

		// get batch
		batch, err := c.GetDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, QueryParams{Offset: offset, Limit: batchSize})
		if err != nil {
			chErr <- err
			return
		}

		// lock for deterministic result manipulation
		lockResult.Lock()

		// (first iteration only) - set totalCount
		if first {
			totalCount = batch.TotalCount
			first = false
		}

		// process batch by calling the provided function
		abort, err := processBatch(batch)
		if err != nil {
			chErr <- err
		}
		if abort {
			close(chAbort)
		}

		// unlock
		lockResult.Unlock()
	}

	// execute first batch sequentially, so that we know the total count before triggering any go-routine
	doProcessBatch(0)

	// determine the total number of calls (considering that we have already performed the first call)
	numCalls := totalCount / batchSize
	if (totalCount % batchSize) == 0 {
		numCalls--
	}

	// process batches concurrently
	for i := 0; i < numCalls; i++ {
		acquire()
		go doProcessBatch(i + 1)
	}

	// func that will close wait channel when all go-routines complete their execution
	go func() {
		wg.Wait()
		close(chWait)
	}()

	// Block until all workers finish their work, keeping track of errors
	for {
		select {
		case err = <-chErr:
			abort = true
		case <-chAbort:
			abort = true
		case <-chWait:
			return
		}
	}
}
