package cantabular

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/stream/jsonstream"
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
	header := createCSVHeader(table.Dimensions)
	if err := write(header); err != nil {
		return nil, fmt.Errorf("error writing the csv header: %w", err)
	}

	// Obtain the CSV rows according to the cantabular dimensions and counts
	for i, count := range table.Values {
		row := createCSVRow(table.Dimensions, i, count)
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

// GraphqlJSONToCSV converts a JSON response in r to CSV on w
// if an error happens, the process is aborted and the error is returned.
func GraphqlJSONToCSV(r io.Reader, w io.Writer) error {
	dec := jsonstream.New(r)
	isStartObj, err := dec.StartObjectComposite()
	if err != nil {
		return fmt.Errorf("error decoding start of json object: %w", err)
	}

	if !isStartObj {
		return errors.New("no json object found in response")
	}
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "data":
			isStartObj, err := dec.StartObjectComposite()
			if err != nil {
				return fmt.Errorf("error decoding start of json object for 'data': %w", err)
			}
			if isStartObj {
				if err := decodeDataFields(dec, w); err != nil {
					return fmt.Errorf("error decoding data fields: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return fmt.Errorf("error decoding end of json object for 'data': %w", err)
				}
			}
		case "errors":
			if err := decodeErrors(dec); err != nil {
				return err
			}
		}
	}
	if err := dec.EndComposite(); err != nil {
		return fmt.Errorf("error decoding end of json object: %w", err)
	}
	return nil
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// If no table cell values are present then no output is written.
func decodeTableFields(dec jsonstream.Decoder, w io.Writer) error {
	var dims Dimensions
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				return fmt.Errorf("error decoding dimensions: %w", err)
			}
		case "error":
			errMsg, err := dec.DecodeString()
			if err != nil {
				return fmt.Errorf("error decoding error message: %w", err)
			}
			if errMsg != nil {
				return fmt.Errorf("table blocked: %s", *errMsg)
			}
		case "values":
			if dims == nil {
				return errors.New("values received before dimensions")
			}
			isStartArray, err := dec.StartArrayComposite()
			if err != nil {
				return fmt.Errorf("error decoding start of json array for 'values': %w", err)
			}
			if isStartArray {
				if err := decodeValues(dec, dims, w); err != nil {
					return fmt.Errorf("error decoding values: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return fmt.Errorf("error decoding end of json array for 'values': %w", err)
				}
			}
		}
	}
	return nil
}

// decodeErrors decodes the errors part of the GraphQL response and
// returns any error with the fount message(s) if there are any.
func decodeErrors(dec jsonstream.Decoder) error {
	var graphqlErrs []struct{ Message string }
	if err := dec.Decode(&graphqlErrs); err != nil {
		return fmt.Errorf("error decoding graphql error response; %w", err)
	}
	var sb strings.Builder
	for _, err := range graphqlErrs {
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(err.Message)
	}
	if sb.Len() > 0 {
		return fmt.Errorf("error(s) returned by graphQL query: %w", errors.New(sb.String()))
	}
	return nil
}

// decodeDataFields decodes the fields of the data part of the GraphQL response, writing CSV to w
func decodeDataFields(dec jsonstream.Decoder, w io.Writer) error {
	var matchName = func(name string) error {
		gotName, err := dec.DecodeName()
		if err != nil {
			return fmt.Errorf("error decoding name: %w", err)
		}
		if gotName != name {
			return fmt.Errorf("expected %q but got %q", name, gotName)
		}
		return nil
	}

	if err := matchName("dataset"); err != nil {
		return fmt.Errorf("failed to match dataset: %w", err)
	}

	isStartObj, err := dec.StartObjectComposite()
	if err != nil {
		return fmt.Errorf("error decoding start of json object composite for 'dataset' value: %w", err)
	}
	if !isStartObj {
		return errors.New(`dataset object expected but "null" found`)
	}

	if err := matchName("table"); err != nil {
		return fmt.Errorf("failed to match table: %w", err)
	}

	isStartObj, err = dec.StartObjectComposite()
	if err != nil {
		return fmt.Errorf("error decoding start of json object for 'table': %w", err)
	}

	if isStartObj {
		if err := decodeTableFields(dec, w); err != nil {
			return fmt.Errorf("error decoding table fields: %w", err)
		}
		if err := dec.EndComposite(); err != nil {
			return fmt.Errorf("error decoding end of json object for 'table': %w", err)
		}
	}
	if err := dec.EndComposite(); err != nil {
		return fmt.Errorf("error decoding end of json object for 'dataset': %w", err)
	}
	return nil
}

// decodeValues decodes the values of the cells in the table, writing CSV to w.
func decodeValues(dec jsonstream.Decoder, dims Dimensions, w io.Writer) (err error) {
	cw := csv.NewWriter(w)

	// csv.Writer errors are sticky, so we only need to check when flushing at the end
	defer func() {
		cw.Flush()
		if cwErr := cw.Error(); cwErr != nil {
			err = fmt.Errorf("csv writer error: %w", cwErr)
		}
	}()

	// Create and write header separately
	header := createCSVHeader(dims)
	if err = cw.Write(header); err != nil {
		return fmt.Errorf("error writing the csv header: %w", err)
	}

	// Obtain the CSV rows according to the cantabular dimensions and counts
	for ti := dims.NewIterator(); dec.More(); {
		count, err := dec.DecodeNumber()
		if err != nil {
			return fmt.Errorf("error decoding count: %w", err)
		}
		row := ti.createCSVRow(dims, count.String())
		if err = cw.Write(row); err != nil {
			return fmt.Errorf("error writing a csv row: %w", err)
		}
		ti.Next()
	}

	return nil
}

// createCSVHeader creates an array of strings corresponding to a csv header
// where each column contains the value of the corresponding dimension, with the last column being the 'count'
func createCSVHeader(dims Dimensions) []string {
	header := make([]string, len(dims)+1)
	for i, dim := range dims {
		header[i+1] = dim.Variable.Label
	}
	header[0] = "count"

	return header
}

// createCSVRow creates an array of strings corresponding to a csv row
// for the provided array of dimension, index and count
// it assumes that the values are sorted with lower weight for the last dimension and higher weight for the first dimension.
func createCSVRow(dims []Dimension, index, count int) []string {
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

// createCSVRow creates an array of strings corresponding to a csv row
// for the provided dimensions and count, using the pointer receiver iterator to determine the column value
func (ti *Iterator) createCSVRow(dims []Dimension, count string) []string {
	row := make([]string, len(dims)+1)
	for i := range dims {
		row[i+1] = ti.CategoryAtColumn(i).Label
	}
	row[0] = count

	return row
}
