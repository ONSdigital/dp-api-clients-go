package cantabular

import (
	"bufio"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseTable(t *testing.T) {

	Convey("Given a Cantabular client", t, func() {
		var c Client

		Convey("When ParseTable is triggered with a valid table", func() {
			resp := testStaticDatasetQuery()
			reader, err := c.ParseTable(resp.Dataset.Table)

			Convey("Then the expected reader is returned without error", func() {
				So(err, ShouldBeNil)
				validateLines(reader, []string{
					"count,City,Number of siblings (3 mappings),Sex",
					"2,London,No siblings,Male",
					"0,London,No siblings,Female",
					"1,London,1 or 2 siblings,Male",
					"3,London,1 or 2 siblings,Female",
					"5,London,3 or more siblings,Male",
					"4,London,3 or more siblings,Female",
					"7,Liverpool,No siblings,Male",
					"6,Liverpool,No siblings,Female",
					"11,Liverpool,1 or 2 siblings,Male",
					"10,Liverpool,1 or 2 siblings,Female",
					"9,Liverpool,3 or more siblings,Male",
					"13,Liverpool,3 or more siblings,Female",
					"14,Belfast,No siblings,Male",
					"12,Belfast,No siblings,Female",
					"16,Belfast,1 or 2 siblings,Male",
					"17,Belfast,1 or 2 siblings,Female",
					"15,Belfast,3 or more siblings,Male",
					"8,Belfast,3 or more siblings,Female",
				})
			})
		})
	})
}

// validateLines scans the provided reader, line by line, and compares with the corresponding line in the provided array.
// It also checks that all the expected lines are present in the reader.
func validateLines(reader *bufio.Reader, expectedLines []string) {
	i := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		So(scanner.Text(), ShouldEqual, expectedLines[i])
		i++
	}
	So(expectedLines, ShouldHaveLength, i) // Check that there aren't any more expected lines
	So(scanner.Err(), ShouldBeNil)
}

// testStaticDatasetQueryResponse returns a valid cantabular StaticDatasetQuery for testing
func testStaticDatasetQuery() *StaticDatasetQuery {
	return &StaticDatasetQuery{
		Dataset: StaticDataset{
			Table: Table{
				Dimensions: []Dimension{
					{
						Variable: VariableBase{
							Name:  "city",
							Label: "City",
						},
						Count: 3,
						Categories: []Category{
							{
								Code:  "0",
								Label: "London",
							},
							{
								Code:  "1",
								Label: "Liverpool",
							},
							{
								Code:  "2",
								Label: "Belfast",
							},
						},
					},
					{
						Variable: VariableBase{
							Name:  "siblings_3",
							Label: "Number of siblings (3 mappings)",
						},
						Count: 3,
						Categories: []Category{
							{
								Code:  "0",
								Label: "No siblings",
							},
							{
								Code:  "1-2",
								Label: "1 or 2 siblings",
							},
							{
								Code:  "3+",
								Label: "3 or more siblings",
							},
						},
					},
					{
						Variable: VariableBase{
							Name:  "sex",
							Label: "Sex",
						},
						Count: 2,
						Categories: []Category{
							{
								Code:  "0",
								Label: "Male",
							},
							{
								Code:  "1",
								Label: "Female",
							},
						},
					},
				},
				Values: []int{2, 0, 1, 3, 5, 4, 7, 6, 11, 10, 9, 13, 14, 12, 16, 17, 15, 8},
			},
		},
	}
}
