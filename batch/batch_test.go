package batch

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var testETag = "testETag"

var (
	errGetter    = errors.New("BatchGetter error")
	errProcessor = errors.New("BatchProcessor error")
)

func TestProcessInConcurrentBatches(t *testing.T) {

	Convey("Given a full slice of 10 items", t, func() {
		full := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

		// batch getter mock generator and calls tracker
		batchGetterCalls := []int{}
		batchGetter := func(batchSize int, retErrorByCall []error) GenericBatchGetter {
			return func(offset int) (interface{}, int, string, error) {
				numCall := len(batchGetterCalls)
				batchGetterCalls = append(batchGetterCalls, offset)
				end := Min(offset+batchSize, len(full))
				batch := full[offset:end]
				return batch, len(full), testETag, retErrorByCall[numCall]
			}
		}

		// batch processor mock generator and calls tracker
		batchProcessorCalls := []interface{}{}
		eTags := []string{}
		batchProcessor := func(retAbortByCall []bool, retErrorByCall []error) GenericBatchProcessor {
			return func(batch interface{}, batchETag string) (abort bool, err error) {
				numCall := len(batchProcessorCalls)
				batchProcessorCalls = append(batchProcessorCalls, batch)
				eTags = append(eTags, batchETag)
				return retAbortByCall[numCall], retErrorByCall[numCall]
			}
		}

		Convey("A batch size of 5 and successful batch getter and processor mocks", func() {
			batchSize := 5
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil, nil})
			processor := batchProcessor([]bool{false, false}, []error{nil, nil})

			Convey("Then processing in batches results in the methods being called twice with the expected offsets and batches", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(batchGetterCalls, ShouldResemble, []int{0, 5})
				So(batchProcessorCalls, ShouldResemble, []interface{}{
					[]string{"0", "1", "2", "3", "4"},
					[]string{"5", "6", "7", "8", "9"}})
				So(eTags, ShouldResemble, []string{testETag, testETag})
			})
		})

		Convey("A batch size of 5 and successful batch getter and processor mocks and 2 workers", func() {
			batchSize := 5
			maxWorkers := 2
			getter := batchGetter(batchSize, []error{nil, nil})
			processor := batchProcessor([]bool{false, false}, []error{nil, nil})

			Convey("Then processing in batches results in the methods being called twice with the expected offsets and batches", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(batchGetterCalls, ShouldResemble, []int{0, 5})
				So(batchProcessorCalls, ShouldResemble, []interface{}{
					[]string{"0", "1", "2", "3", "4"},
					[]string{"5", "6", "7", "8", "9"}})
				So(eTags, ShouldResemble, []string{testETag, testETag})
			})
		})

		Convey("A batch size of 3 and successful batch getter and processor mocks", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil, nil, nil, nil})
			processor := batchProcessor([]bool{false, false, false, false}, []error{nil, nil, nil, nil})

			Convey("Then processing in batches results in the methods being called 4 times with the expected offsets and batches", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(batchGetterCalls, ShouldResemble, []int{0, 3, 6, 9})
				So(batchProcessorCalls, ShouldResemble, []interface{}{
					[]string{"0", "1", "2"},
					[]string{"3", "4", "5"},
					[]string{"6", "7", "8"},
					[]string{"9"}})
				So(eTags, ShouldResemble, []string{testETag, testETag, testETag, testETag})
			})
		})

		Convey("A batch size of 3 and a batch getter that returns error in the first call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{errGetter})
			processor := batchProcessor([]bool{}, []error{})

			Convey("Then processing in batches results in the process being aborted and the same error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldResemble, errGetter)
				So(batchGetterCalls, ShouldResemble, []int{0})           // first call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{}) // no call
				So(eTags, ShouldResemble, []string{})
			})
		})

		Convey("A batch size of 3 and a batch processor that returns error in the first call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil})
			processor := batchProcessor([]bool{false}, []error{errProcessor})

			Convey("Then processing in batches results in the process being aborted and the same error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldResemble, errProcessor)
				So(batchGetterCalls, ShouldResemble, []int{0})         // first call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{ // first call only
					[]string{"0", "1", "2"}})
				So(eTags, ShouldResemble, []string{testETag})
			})
		})

		Convey("A batch size of 3 and a batch processor that aborts the operation (without error) in the first call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil})
			processor := batchProcessor([]bool{true}, []error{nil})

			Convey("Then processing in batches results in the process being aborted and nil error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(batchGetterCalls, ShouldResemble, []int{0})         // first call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{ // first call only
					[]string{"0", "1", "2"}})
				So(eTags, ShouldResemble, []string{testETag})
			})
		})

		Convey("A batch size of 3 and a batch getter that returns error in the third call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil, nil, errGetter})
			processor := batchProcessor([]bool{false, false}, []error{nil, nil})

			Convey("Then processing in batches results in the process being aborted and the same error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldResemble, errGetter)
				So(batchGetterCalls, ShouldResemble, []int{0, 3, 6})   // 3 call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{ // 2 calls only
					[]string{"0", "1", "2"},
					[]string{"3", "4", "5"}})
				So(eTags, ShouldResemble, []string{testETag, testETag})
			})
		})

		Convey("A batch size of 3 and a batch processor that returns error in the third call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil, nil, nil})
			processor := batchProcessor([]bool{false, false, false}, []error{nil, nil, errProcessor})

			Convey("Then processing in batches results in the process being aborted and the same error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldResemble, errProcessor)
				So(batchGetterCalls, ShouldResemble, []int{0, 3, 6})   // 3 call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{ // 3 calls only
					[]string{"0", "1", "2"},
					[]string{"3", "4", "5"},
					[]string{"6", "7", "8"}})
				So(eTags, ShouldResemble, []string{testETag, testETag, testETag})
			})
		})

		Convey("A batch size of 3 and a batch processor that aborts the operation (without error) in the third call", func() {
			batchSize := 3
			maxWorkers := 1
			getter := batchGetter(batchSize, []error{nil, nil, nil})
			processor := batchProcessor([]bool{false, false, true}, []error{nil, nil, nil})

			Convey("Then processing in batches results in the process being aborted and nil error being returned", func() {
				err := ProcessInConcurrentBatches(getter, processor, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(batchGetterCalls, ShouldResemble, []int{0, 3, 6})   // 3 call only
				So(batchProcessorCalls, ShouldResemble, []interface{}{ // 3 calls only
					[]string{"0", "1", "2"},
					[]string{"3", "4", "5"},
					[]string{"6", "7", "8"}})
				So(eTags, ShouldResemble, []string{testETag, testETag, testETag})
			})
		})
	})
}

func TestProcessInBatches(t *testing.T) {

	Convey("Given an array of 10 items and a mock chunk processor function", t, func() {
		items := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
		processedChunks := [][]string{}
		processor := func(chunk []string) error {
			processedChunks = append(processedChunks, chunk)
			return nil
		}

		Convey("Then processing in chunks of size 5 results in the function being called twice with the expected chunks", func() {
			numChunks, err := ProcessInBatches(items, processor, 5)
			So(err, ShouldBeNil)
			So(numChunks, ShouldEqual, 2)
			So(processedChunks, ShouldResemble, [][]string{
				{"0", "1", "2", "3", "4"},
				{"5", "6", "7", "8", "9"}})
		})

		Convey("Then processing in chunks of size 3 results in the function being called four times with the expected chunks, the last one being containing the remaining items", func() {
			numChunks, err := ProcessInBatches(items, processor, 3)
			So(err, ShouldBeNil)
			So(numChunks, ShouldEqual, 4)
			So(processedChunks, ShouldResemble, [][]string{
				{"0", "1", "2"},
				{"3", "4", "5"},
				{"6", "7", "8"},
				{"9"}})
		})
	})
}
