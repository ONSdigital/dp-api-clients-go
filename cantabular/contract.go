package cantabular

import "github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"

// ErrorResponse models the error response from cantabular
type ErrorResponse struct {
	Message string `json:"message"`
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
}

// StaticDatasetQuery holds the query for a static dataset landing page from
// POST [cantabular-ext]/graphql.
// It is used both as the internal query request to GraphQL as well as the
// response to the caller, as GraphQL query responses are essentially
// unmarshalled into the requests.
type StaticDatasetQuery struct {
	Dataset StaticDataset `json:"dataset" graphql:"dataset(name: $name)"`
}

// GetDimensionsResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a ruleBase query to obtain static dataset variables, with 1-level relationships
type GetDimensionsResponse struct {
	Dataset gql.DatasetVariables `json:"dataset"`
}

// GetDimensionOptionsResponse holds the response body for
// POST [cantabular-ext]/graphql
// with a query to obtain static dataset variables and categories, without values.
type GetDimensionOptionsResponse struct {
	Dataset StaticDatasetDimensionOptions `json:"dataset"`
}
