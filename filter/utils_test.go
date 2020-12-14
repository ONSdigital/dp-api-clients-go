package filter

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessInBatches(t *testing.T) {

	Convey("Given an array of 10 items and a mock chunk processor function", t, func() {
		items := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
		processedChunks := [][]string{}
		processor := func(chunk []string) error {
			processedChunks = append(processedChunks, chunk)
			return nil
		}

		Convey("Then processing in chunks of size 5 results in the function being called twice with the expected chunks", func() {
			numChunks, err := processInBatches(items, processor, 5)
			So(err, ShouldBeNil)
			So(numChunks, ShouldEqual, 2)
			So(processedChunks, ShouldResemble, [][]string{
				{"0", "1", "2", "3", "4"},
				{"5", "6", "7", "8", "9"}})
		})

		Convey("Then processing in chunks of size 3 results in the function being called four times with the expected chunks, the last one being containing the remaining items", func() {
			numChunks, err := processInBatches(items, processor, 3)
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
