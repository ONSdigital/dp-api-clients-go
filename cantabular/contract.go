package cantabular

import "github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"

// ErrorResponse models the error response from cantabular
type ErrorResponse struct {
	Message string `json:"message"`
}

type PaginationParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginationResponse struct {
	PaginationParams
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

// GetCodebookRequest holds the query parameters for
// GET [cantabular-srv]/codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookRequest struct {
	DatasetName string
	Variables   []string
	Categories  bool
}

// GetCodebookResponse holds the response body for
// GET [cantabular-srv]/codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookResponse struct {
	Codebook Codebook `json:"codebook"`
	Dataset  Dataset  `json:"dataset"`
}

// StaticDatasetQueryRequest holds the request variables required from the
// caller for making a request for a static dataset landing page from
// POST [cantabular-ext]/graphql
type StaticDatasetQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Variables []string `json:"variables"`
	Filters   []Filter `json:"filters"`
}

// StaticDatasetQuery holds the query for a static dataset landing page from
// POST [cantabular-ext]/graphql.
// It is used both as the internal query request to GraphQL as well as the
// response to the caller, as GraphQL query responses are essentially
// unmarshalled into the requests.
type StaticDatasetQuery struct {
	Dataset StaticDataset `json:"dataset" graphql:"dataset(name: $name)"`
}

// GetDimensionsByNameRequest holds the request variables required from the
// caller for making a request to obtain dimensions (Cantabular variables) by name
// POST [cantabular-ext]/graphql
type GetDimensionsByNameRequest struct {
	Dataset        string
	DimensionNames []string
}

type GetDimensionsRequest struct {
	PaginationParams
	Dataset string
	Text    string
}

// SearchDimensionsRequest holds the request variables required from the
// caller for making a request to search dimensions (Cantabular variables) by text
// POST [cantabular-ext]/graphql
type SearchDimensionsRequest struct {
	Dataset string
	Text    string
}

// GetDimensionsResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a query to obtain variables
type GetDimensionsResponse struct {
	PaginationResponse
	Dataset gql.Dataset `json:"dataset"`
}

// GetGeographyDimensionsRequest holds the request parameters for
// POST [cantabular-ext]/graphql
// with a query to obtain geography variables
type GetGeographyDimensionsRequest struct {
	PaginationParams
	Dataset string `json:"dataset"`
}

// GetGeographyDimensionsResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a query to obtain geography variables
type GetGeographyDimensionsResponse struct {
	PaginationResponse
	Dataset gql.Dataset `json:"dataset"`
}

// GetDimensionOptionsRequest holds the request variables required from the
// caller for making a request to obtain dimension options (categories)
// for the provided cantabular Dataset and dimension names (Cantabular variables)
//
// POST [cantabular-ext]/graphql with the encoded query
type GetDimensionOptionsRequest struct {
	Dataset        string
	DimensionNames []string
	Filters        []Filter
}

// GetDimensionOptionsResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a query to obtain static dataset variables and categories, without values.
type GetDimensionOptionsResponse struct {
	Dataset StaticDatasetDimensionOptions `json:"dataset"`
}

// GetAreasRequest holds the request variables required for the
// POST [cantabular-ext]/graphql QueryAreas query.
type GetAreasRequest struct {
	PaginationParams
	Dataset  string
	Variable string
	Category string
}

// GetAreasResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a query to obtain static dataset variables and categories, without values.
type GetAreasResponse struct {
	Dataset gql.Dataset `json:"dataset"`
}

// GetParentsRequest holds the input parameters for the GetParents query
type GetParentsRequest struct {
	Dataset  string
	Variable string
}

// GetParentsResponse is the response body for the GetParents query
type GetParentsResponse struct {
	Dataset gql.Dataset `json:"dataset"`
}

// GetParentAreaCountRequest holds the input parameters for the GetParents query
type GetParentAreaCountRequest struct {
	Dataset  string
	Variable string
	Parent   string
	Codes    []string
}

// GetParentAreaCountResponse is the response body for the GetParentAreaCount query
type GetParentAreaCountResponse struct {
	Dataset struct {
		Table Table `json:"table`
	} `json:"dataset"`
}

// GetParentAreaCountResult is the useful part of the response for GetParentAreaCount
type GetParentAreaCountResult struct {
	Dimension Dimension
}
