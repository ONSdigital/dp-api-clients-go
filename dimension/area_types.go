package dimension

// AreaType is an area-type model with ID and Label
type AreaType struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// GetAreaTypesResponse is the response object for GET /area-types
type GetAreaTypesResponse struct {
	AreaTypes []AreaType `json:"area-types"`
}
