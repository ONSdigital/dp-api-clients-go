package batch

import (
	"sync"
	"time"
)

// GenericBatchGetter defines the method signature for a batch getter to obtain a batch of some generic resource
type GenericBatchGetter func(offset int) (batch interface{}, totalCount int, eTag string, err error)

// GenericBatchProcessor defines the method signature for a batch processor to process a batch of some generic resource
type GenericBatchProcessor func(batch interface{}, batchETag string) (abort bool, err error)

// ProcessInConcurrentBatches is a generic method to concurrently obtain some resource in batches and then process each batch
func ProcessInConcurrentBatches(getBatch GenericBatchGetter, processBatch GenericBatchProcessor, batchSize, maxWorkers int) (err error) {
	wg := sync.WaitGroup{}
	chWait := make(chan struct{})
	chErr := make(chan error, maxWorkers)
	chAbort := make(chan struct{})
	chSemaphore := make(chan struct{}, maxWorkers)

	lockResult := sync.Mutex{}

	// worker add delta to workers WaitGroup and acquire semaphore
	acquire := func() {
		wg.Add(1)
		chSemaphore <- struct{}{}
	}

	// worker release semaphore and workers WaitGroup delta
	release := func() {
		<-chSemaphore
		wg.Done()
	}

	// abort closes the abort channel if it's not already closed
	abort := func() {
		select {
		case <-chAbort:
		default:
			close(chAbort)
		}
	}

	// isAborting returns true if the abort channel is closed
	isAborting := func() bool {
		select {
		case <-chAbort:
			return true
		default:
			return false
		}
	}

	// func executed in each go-routine to process the batch and send errors to the error channel
	doProcessBatch := func(offset int) {
		defer release()

		// Abort if needed
		if isAborting() {
			return
		}

		// get batch
		batch, _, batchETag, err := getBatch(offset)
		if err != nil {
			chErr <- err
			abort()
			return
		}

		// lock to prevent concurrent result manipulation
		lockResult.Lock()
		defer lockResult.Unlock()

		// process batch by calling the provided function
		forceAbort, err := processBatch(batch, batchETag)
		if err != nil {
			chErr <- err
			abort()
		}
		if forceAbort {
			abort()
		}
	}

	// get first batch sequentially, so that we know the total count before triggering any further go-routine
	batch, totalCount, batchETag, err := getBatch(0)
	if err != nil {
		return err
	}

	// process first batch by calling the provided function
	forceAbort, err := processBatch(batch, batchETag)
	if forceAbort || err != nil {
		return err
	}

	// determine the total number of remaining calls, considering that we have already performed the first one
	numCalls := totalCount / batchSize
	if (totalCount % batchSize) == 0 {
		numCalls--
	}

	// process remaining batches concurrently
	for i := 0; i < numCalls; i++ {
		acquire()
		go doProcessBatch((i + 1) * batchSize)
	}

	// func that will close wait channel when all go-routines complete their execution
	go func() {
		wg.Wait()
		time.Sleep(time.Millisecond)
		close(chWait)
	}()

	// Block until all workers finish their work, keeping track of errors
	for {
		select {
		case err = <-chErr:
		case <-chWait:
			return err
		}
	}
}

// ProcessInBatches is a generic method that splits the provided items in batches and calls processBatch for each batch
func ProcessInBatches(items []string, processBatch func([]string) error, batchSize int) (processedBatches int, err error) {
	// Get batch splits for provided items
	numFullChunks := len(items) / batchSize
	remainingSize := len(items) % batchSize
	processedBatches = numFullChunks

	// process full batches
	for i := 0; i < numFullChunks; i++ {
		chunk := items[i*batchSize : (i+1)*batchSize]
		if err := processBatch(chunk); err != nil {
			return i, err
		}
	}

	// process any remaining
	if remainingSize > 0 {
		processedBatches = numFullChunks + 1
		lastChunk := items[numFullChunks*batchSize : (numFullChunks*batchSize + remainingSize)]
		if err := processBatch(lastChunk); err != nil {
			return numFullChunks, err
		}
	}
	return processedBatches, nil
}

// Min returns the lowest value
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
