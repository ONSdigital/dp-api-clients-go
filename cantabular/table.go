package cantabular

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/jsonstream"
)

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
	mustMatchName := func(name string) error {
		if gotName := dec.DecodeName(); gotName != name {
			return fmt.Errorf("expected %q but got %q", name, gotName)
		}
		return nil
	}

	if err := mustMatchName("dataset"); err != nil {
		return fmt.Errorf("failed to match dataset: %w", err)
	}
	if !dec.StartObjectComposite() {
		return errors.New(`dataset object expected but "null" found`)
	}

	if err := mustMatchName("table"); err != nil {
		return fmt.Errorf("failed to match table: %w", err)
	}
	if dec.StartObjectComposite() {
		decodeTableFields(dec, w)
		dec.EndComposite()
	}
	dec.EndComposite()
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
		row := createCSVRow(ti, dims, dec.DecodeNumber().String())
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
// for the provided iterator of values, dimensions and count
func createCSVRow(ti *Iterator, dims []Dimension, count string) []string {
	row := make([]string, len(dims)+1)
	for i := range dims {
		row[i+1] = ti.CategoryAtColumn(i).Label
	}
	row[0] = count

	return row
}
