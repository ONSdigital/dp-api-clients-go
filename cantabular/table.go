package cantabular

import (
	"bufio"
	"bytes"
	"context"
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
	Values     []float32   `json:"values"`
	Error      string      `json:"error,omitempty" `
}

// possible errors that may be returned while parsing a 'data' field value
// that, if present, need to be reported under the 'errors' field
var (
	errNullDataset = errors.New(`dataset object expected but "null" found`)
	errNullTable   = errors.New(`table object expected but "null" found`)
)

// ParseTable takes a table from a GraphQL response and parses it into a
// header and rows of counts (observations) ready to be read line-by-line.
func (c *Client) ParseTable(table Table) (*bufio.Reader, error) {
	// Create CSV writer with underlying buffer
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	// aux func to write to the csv writer, returning any error (returned by w.Write or w.Errors)
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
		row := createCSVRow(table.Dimensions, i, int(count))
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

// GraphQLJSONToCSV converts a JSON response in r to CSV on w, returning the row count
// if an error happens, the process is aborted and the error is returned.
func GraphQLJSONToCSV(ctx context.Context, r io.Reader, w io.Writer) (int32, error) {
	dec := jsonstream.New(r)

	// errData represents a possible error that may be returned by 'decodeDataFields',
	// as long as it is reported in 'errors'
	var errData error

	// rowCount is the number of CSV rows that are decoded from r
	var rowCount int32

	// find starting '{'
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return 0, fmt.Errorf("error decoding start of json object: %w", err)
	} else if !isStartObj {
		return 0, errors.New("no json object found in response")
	}

	// decode 'data' and 'error' fields
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return 0, fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "data":
			if rowCount, err = decodeDataFields(ctx, dec, w); err != nil {
				// null values for 'dataste' or 'table' are ok as long as the error is reported under the 'errors' field
				if err == errNullDataset || err == errNullTable {
					errData = err
					break
				}
				return 0, err
			}
		case "errors":
			if err := decodeErrors(dec); err != nil {
				return 0, err
			}
		}
	}

	// find final '}'
	if err := dec.EndComposite(); err != nil {
		return 0, fmt.Errorf("error decoding end of json object: %w", err)
	}

	// check if there was an error in the "data" section that was not reported in the "error" section
	if errData != nil {
		return 0, fmt.Errorf("error found parsing 'data' filed, but no error was reported in 'error' filed: %w", errData)
	}
	return rowCount, nil
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// It returns the total number of rows, including the header.
// If no table cell values are present then no output is written.
func decodeTableFields(ctx context.Context, dec jsonstream.Decoder, w io.Writer) (rowCount int32, err error) {
	var dims Dimensions
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return 0, fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				return 0, fmt.Errorf("error decoding dimensions: %w", err)
			}
		case "error":
			errMsg, err := dec.DecodeString()
			if err != nil {
				return 0, fmt.Errorf("error decoding error message: %w", err)
			}
			if errMsg != nil {
				return 0, fmt.Errorf("table blocked: %s", *errMsg)
			}
		case "values":
			if dims == nil {
				return 0, errors.New("values received before dimensions")
			}
			isStartArray, err := dec.StartArrayComposite()
			if err != nil {
				return 0, fmt.Errorf("error decoding start of json array for 'values': %w", err)
			}
			if isStartArray {
				if rowCount, err = decodeValues(ctx, dec, dims, w); err != nil {
					return 0, fmt.Errorf("error decoding values: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return 0, fmt.Errorf("error decoding end of json array for 'values': %w", err)
				}
			}
		}
	}
	return rowCount, nil
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
// if expects to find the following nested values: {"dataset": {"table": {...}}}
// it propagates the row count returned by 'decodeTableFields'
// the end-composite values are always decoded according to the reached depth
func decodeDataFields(ctx context.Context, dec jsonstream.Decoder, w io.Writer) (rowCount int32, err error) {
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

	depth := 0

	// decode endComposites according to reached json depth
	defer func() {
		for i := 0; i < depth; i++ {
			if e := dec.EndComposite(); e != nil {
				err = fmt.Errorf("error decoding end of json object: %w", e)
			}
		}
	}()

	// find starting '{' for 'data' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return 0, fmt.Errorf("error decoding start of json object for 'data': %w", err)
	} else if !isStartObj {
		return 0, nil // no value for 'data'
	}
	depth++

	// find 'dataset' key
	if err := matchName("dataset"); err != nil {
		return 0, fmt.Errorf("failed to match dataset: %w", err)
	}

	// find '{' for 'dataset' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return 0, fmt.Errorf("error decoding start of json object composite for 'dataset' value: %w", err)
	} else if !isStartObj {
		return 0, errNullDataset // valid scenario if 'data' is parsed before 'errors'
	}
	depth++

	// find 'table' key
	if err := matchName("table"); err != nil {
		return 0, fmt.Errorf("failed to match table: %w", err)
	}

	// find '{' for 'table' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return 0, fmt.Errorf("error decoding start of json object for 'table': %w", err)
	} else if !isStartObj {
		return 0, errNullTable // valid scenario if 'data' is parsed before 'errors'
	}
	depth++

	// Decode table fields
	if rowCount, err = decodeTableFields(ctx, dec, w); err != nil {
		return 0, fmt.Errorf("error decoding table fields: %w", err)
	}
	return rowCount, nil
}

// decodeValues decodes the values of the cells in the table, writing CSV to w.
// It returns the total number of rows, including the header.
func decodeValues(ctx context.Context, dec jsonstream.Decoder, dims Dimensions, w io.Writer) (rowCount int32, err error) {
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
		return 0, fmt.Errorf("error writing the csv header: %w", err)
	}
	rowCount = 1 // number of rows, including headers

	// Obtain the CSV rows according to the cantabular dimensions and counts
	for ti := dims.NewIterator(ctx); dec.More(); {
		count, err := dec.DecodeNumber()
		if err != nil {
			return 0, fmt.Errorf("error decoding count: %w", err)
		}
		row, err := ti.createCSVRow(dims, count.String())
		if err != nil {
			return 0, fmt.Errorf("error parsing a csv row: %w", err)
		}
		if err = cw.Write(row); err != nil {
			return 0, fmt.Errorf("error writing a csv row: %w", err)
		}
		rowCount++
		if err := ti.Next(); err != nil {
			return 0, fmt.Errorf("error iterating to next row: %w", err)
		}
	}

	return rowCount, nil
}

// createCSVHeader creates an array of strings corresponding to a csv header
// where each column contains the value of the corresponding dimension, with the last column being the 'count'
func createCSVHeader(dims Dimensions) []string {
	i := 0
	l := len(dims) * 2
	header := make([]string, l+1)
	for _, dim := range dims {
		header[i] = fmt.Sprintf("%s Code", dim.Variable.Label)
		i++
		header[i] = dim.Variable.Label
		i++
	}
	header[l] = "Observation"

	return header
}

// createCSVRow creates an array of strings corresponding to a csv row
// for the provided array of dimension, index and count
// it assumes that the values are sorted with lower weight for the last dimension and higher weight for the first dimension.
func createCSVRow(dims []Dimension, index, count int) []string {
	l := len(dims)
	row := make([]string, l*2+1)

	// Iterate dimensions starting from the last one (lower weight)
	for i := l - 1; i >= 0; i-- {
		catIndex := index % dims[i].Count               // Index of the category for the current dimension
		row[i*2] = dims[i].Categories[catIndex].Code    // The CSV column corresponds to the label of the Category
		row[i*2+1] = dims[i].Categories[catIndex].Label // The CSV column corresponds to the label of the Category
		index /= dims[i].Count                          // Modify index for next iteration
	}

	row[l*2] = fmt.Sprintf("%d", count)

	return row
}

// createCSVRow creates an array of strings corresponding to a csv row
// for the provided dimensions and count, using the pointer receiver iterator to determine the column value
func (it *Iterator) createCSVRow(dims []Dimension, count string) ([]string, error) {
	l := len(dims) * 2
	row := make([]string, l+1)
	for i := range dims {
		category, err := it.CategoryAtColumn(i)
		if err != nil {
			return []string{}, fmt.Errorf("failed to find category at column %d, : %w", i, err)
		}
		row[i*2] = category.Code
		row[i*2+1] = category.Label
	}
	row[l] = count

	return row, nil
}
