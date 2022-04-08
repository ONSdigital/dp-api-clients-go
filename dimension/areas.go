package dimension

// Area is an area model with ID and Label
type Area struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AreaType string `json:"area-type"`
}

// GetAreasResponse is the response object for GET /areas
type GetAreasResponse struct {
	Areas []Area `json:"areas"`
}
