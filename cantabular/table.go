package cantabular

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
)

// Table represents the 'table' field from the GraphQL dataset
// query response
type Table struct {
	Dimensions []Dimension `json:"dimensions"`
	Values     []int       `json:"values"`
	Error      string      `json:"error,omitempty" `
}

// ParseTable takes a table from a GraphQL response and parses it into a
// header and rows of counts (observations) ready to be read line-by-line.
func (c *Client) ParseTable(table Table) (*bufio.Reader, error) {
	// Create CSV writer with underlying buffer
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	// aux func to write to the csv writer, returning any error (returned by w.Write or w.Error)
	write := func(record []string) error {
		if err := w.Write(record); err != nil {
			return err
		}
		return w.Error()
	}

	// Create and write header separately
	header := c.createCSVHeader(table.Dimensions)
	if err := write(header); err != nil {
		return nil, fmt.Errorf("error writing the csv header: %w", err)
	}

	// Obtain the CSV rows according to the cantabular dimensions and counts
	for i, count := range table.Values {
		row := c.createCSVRow(table.Dimensions, i, count)
		if err := write(row); err != nil {
			return nil, fmt.Errorf("error writing a csv row: %w", err)
		}
	}

	// Flush to make sure all data is present in the byte buffer
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("error flushing the csv writer: %w", err)
	}

	// Return a reader with the same underlying Byte buffer as written by the csv writer
	return bufio.NewReader(b), nil
}

// createCSVHeader creates an array of strings corresponding to a csv header
// where each column contains the value of the corresponding dimension, with the last column being the 'count'
func (c *Client) createCSVHeader(dims []Dimension) []string {
	header := make([]string, len(dims)+1)
	for i, dim := range dims {
		header[i+1] = dim.Variable.Label
	}
	header[0] = "cantabular_blob"

	return header
}

// createCSVRow creates an array of strings corresponding to a csv row
// for the provided array of dimension, index and count
// it assumes that the values are sorted with lower weight for the last dimension and higher weight for the first dimension.
func (c *Client) createCSVRow(dims []Dimension, index, count int) []string {
	row := make([]string, len(dims)+1)

	// Iterate dimensions starting from the last one (lower weight)
	for i := len(dims) - 1; i >= 0; i-- {
		catIndex := index % dims[i].Count             // Index of the category for the current dimension
		row[i+1] = dims[i].Categories[catIndex].Label // The CSV column corresponds to the label of the Category
		index /= dims[i].Count                        // Modify index for next iteration
	}

	row[0] = fmt.Sprintf("%d", count)

	return row
}
