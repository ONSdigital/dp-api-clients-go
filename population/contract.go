package population

type PaginationParams struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type PaginationResponse struct {
	PaginationParams
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

type AuthTokens struct {
	UserAuthToken    string
	ServiceAuthToken string
}
