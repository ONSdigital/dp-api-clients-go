package cantabular

// ErrorResponse models the error response from cantabular
type ErrorResponse struct{
	Message string `json:"message"`
	Locations []struct{
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

// GetCodebookRequest holds the query parameters for GET [cantabular-srv]/codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookRequest struct{
	DatasetName string 
	Variables   []string
	Categories  bool
}

// GetCodebookRequest holds the response body for GET [cantabular-srv]/codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookResponse struct{
	Codebook Codebook `json:"codebook"`
	Dataset  Dataset  `json:"dataset"`
}

// StaticDatasetQueryRequest holds the request values for making a request for
// a static dataset landing page from POST [cantabular-ext]/graphql
type StaticDatasetQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Variables []string `json:"variables"`
}

// StaticDatasetQueryRequest holds the response for a static dataset landing
// page from POST [cantabular-ext]/graphql
type StaticDatasetQueryResponse  struct {
	Data struct{
		Dataset StaticDataset
	} `json:"data"`
	Errors []ErrorResponse `json:"errors"`
}
