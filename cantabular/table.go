package cantabular

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/stream/jsonstream"
)

type Categories struct {
	Code  string `json:"code,omitempty"`
	Label string `json:"label,omitempty"`
}
type RuleVariable struct {
	Categories []Categories `json:"categories,omitempty"`
	Count      int          `json:"count,omitempty"`
}

type Rules struct {
	Blocked RuleVariable `json:"blocked,omitempty"`
	Passed  RuleVariable `json:"passed,omitempty"`
	Total   RuleVariable `json:"evaluated,omitempty"`
}

// Table represents the 'table' field from the GraphQL dataset
// query response
type Table struct {
	Dimensions []Dimension `json:"dimensions"`
	Values     []float32   `json:"values"`
	Error      string      `json:"error,omitempty" `
	Rules      Rules       `json:"rules,omitempty"`
}

type ObservationDimension struct {
	Dimension   string `bson:"dimension"           json:"dimension"`
	DimensionID string `bson:"dimension_id"           json:"dimension_id"`
	Option      string `bson:"option"           json:"option"`
	OptionID    string `bson:"option_id"           json:"option_id"`
}

type GetObservationResponse struct {
	Dimensions  []ObservationDimension `bson:"dimensions"           json:"dimensions"`
	Observation float32                `bson:"observation"   json:"observation"`
}

type GetObservationsResponse struct {
	Observations      []GetObservationResponse `bson:"observations"           json:"observations"`
	Links             DatasetJSONLinks         `json:"links"`
	TotalObservations int                      `json:"total_observations"`
	BlockedAreas      int                      `json:"blocked_areas"`
	TotalAreas        int                      `json:"total_areas"`
	AreasReturned     int                      `json:"areas_returned"`
}

type DatasetJSONLinks struct {
	Self Link `bson:"self"       json:"self"`
}

type Link struct {
	HREF  string `bson:"href"            json:"href"`
	Label string `bson:"label,omitempty" json:"label,omitempty"`
	ID    string `bson:"id,omitempty"    json:"id,omitempty"`
}

// possible errors that may be returned while parsing a 'data' field value
// that, if present, need to be reported under the 'errors' field
var (
	errNullDataset = errors.New(`dataset object expected but "null" found`)
	errNullTable   = errors.New(`table object expected but "null" found`)
)

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
			fmt.Println("IN THE DATA BIT")
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

// GraphQLJSONToJson converts a JSON response in r to a different schema on w, returning the response
// if an error happens, the process is aborted and the error is returned.
func GraphQLJSONToJson(ctx context.Context, r io.Reader, w io.Writer) (GetObservationsResponse, error) {
	dec := jsonstream.New(r)

	// errData represents a possible error that may be returned by 'decodeDataFields',
	// as long as it is reported in 'errors'
	var errData error

	var getObservationsResponse GetObservationsResponse

	// find starting '{'
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("error decoding start of json object: %w", err)
	} else if !isStartObj {
		return GetObservationsResponse{}, errors.New("no json object found in response")
	}

	// decode 'data' and 'error' fields
	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return GetObservationsResponse{}, fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "data":
			if getObservationsResponse, err = decodeDataFieldsJson(ctx, dec, w); err != nil {
				// null values for 'dataset' or 'table' are ok as long as the error is reported under the 'errors' field
				if err == errNullDataset || err == errNullTable {
					errData = err
					break
				}
				return GetObservationsResponse{}, err
			}
		case "errors":
			if err := decodeErrors(dec); err != nil {
				fmt.Println(err)
				return GetObservationsResponse{}, err
			}
		}
	}

	// find final '}'
	if err := dec.EndComposite(); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("In GraphQLJSONToJson error decoding end of json object: %w", err)
	}

	// check if there was an error in the "data" section that was not reported in the "error" section
	if errData != nil {
		return GetObservationsResponse{}, fmt.Errorf("error found parsing 'data' filed, but no error was reported in 'error' filed: %w", errData)
	}
	return getObservationsResponse, nil
}

func getDimensionRow(query StaticDatasetQuery, dimIndices []int, dimIndex int) (value []ObservationDimension) {

	var observationDimensions []ObservationDimension

	for index, element := range dimIndices {
		dimension := query.Dataset.Table.Dimensions[index]

		observationDimensions = append(observationDimensions, ObservationDimension{
			Dimension:   dimension.Variable.Label,
			DimensionID: dimension.Variable.Name,
			Option:      dimension.Categories[element].Label,
			OptionID:    dimension.Categories[element].Code,
		})
	}

	return observationDimensions
}

func toGetDatasetObservationsResponse(query StaticDatasetQuery, ctx context.Context) (GetObservationsResponse, error) {

	var getObservationResponse []GetObservationResponse

	dimLength := make([]int, 0)
	dimIndices := make([]int, 0)

	for k := 0; k < len(query.Dataset.Table.Dimensions); k++ {
		dimLength = append(dimLength, len(query.Dataset.Table.Dimensions[k].Categories))
		dimIndices = append(dimIndices, 0)
	}

	for v := 0; v < len(query.Dataset.Table.Values); v++ {
		dimension := getDimensionRow(query, dimIndices, v)
		getObservationResponse = append(getObservationResponse, GetObservationResponse{
			Dimensions:  dimension,
			Observation: query.Dataset.Table.Values[v],
		})

		i := len(dimIndices) - 1
		for i >= 0 {
			dimIndices[i] += 1
			if dimIndices[i] < dimLength[i] {
				break
			}
			dimIndices[i] = 0
			i -= 1
		}

	}

	var getObservationsResponse GetObservationsResponse
	getObservationsResponse.Observations = getObservationResponse
	getObservationsResponse.TotalObservations = len(query.Dataset.Table.Values)

	getObservationsResponse.BlockedAreas = query.Dataset.Table.Rules.Blocked.Count
	getObservationsResponse.TotalAreas = query.Dataset.Table.Rules.Total.Count
	getObservationsResponse.AreasReturned = query.Dataset.Table.Rules.Total.Count

	return getObservationsResponse, nil
}

// decodeTableFields decodes the fields of the table part of the GraphQL response, writing CSV to w.
// It returns the total number of rows, including the header.
// If no table cell values are present then no output is written.
func decodeTableFields(ctx context.Context, dec jsonstream.Decoder, w io.Writer) (rowCount int32, err error) {
	var dims Dimensions
	var rules Rules
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
		case "rules":
			if err := dec.Decode(&rules); err != nil {
				return 0, fmt.Errorf("error decoding rules: %w", err)
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

func decodeTableFieldsJson(ctx context.Context, dec jsonstream.Decoder, w io.Writer) (getObservationsResponse GetObservationsResponse, err error) {
	var dims Dimensions
	var rules Rules
	var getObsResponse GetObservationsResponse
	blockedCount := 0
	totalAreas := 0
	areasReturned := 0

	for dec.More() {
		field, err := dec.DecodeName()
		if err != nil {
			return GetObservationsResponse{}, fmt.Errorf("error decoding field: %w", err)
		}
		switch field {
		case "dimensions":
			if err := dec.Decode(&dims); err != nil {
				return GetObservationsResponse{}, fmt.Errorf("error decoding dimensions: %w", err)
			}
		case "error":
			errMsg, err := dec.DecodeString()
			if err != nil {
				return GetObservationsResponse{}, fmt.Errorf("error decoding error message: %w", err)
			}
			if errMsg != nil {
				return GetObservationsResponse{}, fmt.Errorf("table blocked: %s", *errMsg)
			}
		case "rules":
			if err := dec.Decode(&rules); err != nil {
				return GetObservationsResponse{}, fmt.Errorf("error decoding rules: %w", err)
			}
			blockedCount = rules.Blocked.Count
			areasReturned = rules.Passed.Count
			totalAreas = rules.Total.Count
		case "values":
			if dims == nil {
				return GetObservationsResponse{}, errors.New("values received before dimensions")
			}
			isStartArray, err := dec.StartArrayComposite()
			if err != nil {
				return GetObservationsResponse{}, fmt.Errorf("error decoding start of json array for 'values': %w", err)
			}
			if isStartArray {
				if getObsResponse, err = decodeValuesJson(ctx, dec, dims, w); err != nil {
					return GetObservationsResponse{}, fmt.Errorf("error decoding values: %w", err)
				}
				if err := dec.EndComposite(); err != nil {
					return GetObservationsResponse{}, fmt.Errorf("error decoding end of json array for 'values': %w", err)
				}
			}
		}
	}
	getObsResponse.AreasReturned = areasReturned
	getObsResponse.BlockedAreas = blockedCount
	getObsResponse.TotalAreas = totalAreas
	return getObsResponse, nil
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
				err = fmt.Errorf("In decodeDataFields error decoding end of json object: %w", e)
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

func decodeDataFieldsJson(ctx context.Context, dec jsonstream.Decoder, w io.Writer) (getObservationsResponse GetObservationsResponse, err error) {
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
				err = fmt.Errorf("in decodeDataFieldsJson error decoding end of json object: %w", e)
			}
		}
	}()

	// find starting '{' for 'data' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("error decoding start of json object for 'data': %w", err)
	} else if !isStartObj {
		return GetObservationsResponse{}, nil // no value for 'data'
	}
	depth++

	// find 'dataset' key
	if err := matchName("dataset"); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("failed to match dataset: %w", err)
	}

	// find '{' for 'dataset' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("error decoding start of json object composite for 'dataset' value: %w", err)
	} else if !isStartObj {
		return GetObservationsResponse{}, errNullDataset // valid scenario if 'data' is parsed before 'errors'
	}
	depth++

	// find 'table' key
	if err := matchName("table"); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("failed to match table: %w", err)
	}

	// find '{' for 'table' value
	if isStartObj, err := dec.StartObjectComposite(); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("error decoding start of json object for 'table': %w", err)
	} else if !isStartObj {
		return GetObservationsResponse{}, errNullTable // valid scenario if 'data' is parsed before 'errors'
	}
	depth++

	// Decode table fields
	if getObservationsResponse, err = decodeTableFieldsJson(ctx, dec, w); err != nil {
		return GetObservationsResponse{}, fmt.Errorf("error decoding table fields: %w", err)
	}
	return getObservationsResponse, nil
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

func decodeValuesJson(ctx context.Context, dec jsonstream.Decoder, dims Dimensions, w io.Writer) (getObsResponse GetObservationsResponse, err error) {

	obsResp := json.NewEncoder(w)
	var getObservationsResponse GetObservationsResponse

	// Obtain the Json objects according to the cantabular dimensions and counts
	for ti := dims.NewIterator(ctx); dec.More(); {
		count, err := dec.DecodeNumber()
		if err != nil {
			return GetObservationsResponse{}, fmt.Errorf("error decoding count: %w", err)
		}
		row, err := ti.createJsonObject(dims, count.String())
		if err != nil {
			return GetObservationsResponse{}, fmt.Errorf("error parsing a csv row: %w", err)
		}
		obsResp.Encode(row)
		getObservationsResponse.Observations = append(getObservationsResponse.Observations, row)

		if err := ti.Next(); err != nil {
			return GetObservationsResponse{}, fmt.Errorf("error iterating to next row: %w", err)
		}
	}

	return getObservationsResponse, nil
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

func (it *Iterator) createJsonObject(dims []Dimension, count string) (GetObservationResponse, error) {

	var obsResponse GetObservationResponse

	var obsDimArr []ObservationDimension
	for i := range dims {
		var obsDim ObservationDimension

		category, err := it.CategoryAtColumn(i)

		if err != nil {
			return GetObservationResponse{}, fmt.Errorf("failed to find category at column %d, : %w", i, err)
		}
		obsDim.Dimension = it.dims[i].Variable.Label
		obsDim.DimensionID = it.dims[i].Variable.Name
		obsDim.Option = category.Label
		obsDim.OptionID = category.Code
		obsDimArr = append(obsDimArr, obsDim)

	}

	obsResponse.Dimensions = obsDimArr
	observation, err := strconv.ParseFloat(count, 32)
	if err != nil {
		return GetObservationResponse{}, fmt.Errorf("invalid observation %d, : %w", 0, err)
	}
	obsResponse.Observation = float32(observation)
	return obsResponse, nil
}
