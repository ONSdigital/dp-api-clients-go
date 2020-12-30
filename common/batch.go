package common

import "sync"

// GenericBatchGetter defines the method signature for a batch getter to obtain a batch of some generic resource
type GenericBatchGetter func(offset int) (batch interface{}, totalCount int, err error)

// GenericBatchProcessor defines the method signature for a batch processor to process a batch of some generic resource
type GenericBatchProcessor func(batch interface{}) (abort bool, err error)

// ProcessInConcurrentBatches is a generic method to concurrently obtain some resource in batches and then process each batch
func ProcessInConcurrentBatches(getBatch GenericBatchGetter, processBatch GenericBatchProcessor, batchSize, maxWorkers int) (err error) {
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

	// func executed in each go-routine to process the batch and send errors to the error channel
	doProcessBatch := func(offset int) {
		defer release()

		// Abort if needed
		if abort {
			return
		}

		// get batch
		batch, tc, err := getBatch(offset)
		if err != nil {
			chErr <- err
			return
		}

		// lock for deterministic result manipulation
		lockResult.Lock()

		// (first iteration only) - set totalCount
		if first {
			totalCount = tc
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
