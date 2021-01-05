package common

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
			return func(offset int) (interface{}, int, error) {
				numCall := len(batchGetterCalls)
				batchGetterCalls = append(batchGetterCalls, offset)
				end := Min(offset+batchSize, len(full))
				batch := full[offset:end]
				return batch, len(full), retErrorByCall[numCall]
			}
		}

		// batch processor mock generator and calls tracker
		batchProcessorCalls := []interface{}{}
		batchProcessor := func(retAbortByCall []bool, retErrorByCall []error) GenericBatchProcessor {
			return func(batch interface{}) (abort bool, err error) {
				numCall := len(batchProcessorCalls)
				batchProcessorCalls = append(batchProcessorCalls, batch)
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
			})
		})

	})

}
