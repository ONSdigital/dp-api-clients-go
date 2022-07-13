package population

// Area is an area model with ID and Label
type Area struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AreaType string `json:"area_type"`
}

// GetAreasResponse is the response object for GET /areas
type GetAreasResponse struct {
	Areas []Area `json:"areas"`
}

// Area is an area model with ID and Label
type AreaTypes struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TotalCount int    `json:"total_count"`
}

// GetAreaTypeParentsResponse is the response object for GET /areas
type GetAreaTypeParentsResponse struct {
	AreaTypes []AreaTypes `json:"area_types"`
}
