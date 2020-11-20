package filter

// processInBatches is an aux function that splits the provided items in batches and calls processBatch for each batch
func processInBatches(items []string, processBatch func([]string) error, batchSize int) (processedBatches int, err error) {
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
