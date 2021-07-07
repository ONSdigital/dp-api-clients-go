package cantabular

// ErrorResponse models the error response from cantabular
type ErrorResponse struct{
	Message string `json:"message"`
}

// GetCodebookRequest holds the query parameters for GET /codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookRequest struct{
	DatasetName string 
	Variables   []string
	Categories  bool
}

// GetCodebookRequest holds the response body for GET /codebook/{dataset}?cats=xxx&v=xxx
type GetCodebookResponse struct{
	Codebook Codebook `json:"codebook"`
	Dataset  Dataset  `json:"dataset"`
}
