package population

import "github.com/ONSdigital/dp-api-clients-go/v2/cantabular"

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

// GetAreasResponse is the response object for GET /areas
type GetAreaResponse struct {
	Area Area `json:"area"`
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

// Dimension is an area-type model with ID and Label
type Dimension struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	TotalCount int    `json:"total_count"`
}

// GetDimensionsResponse is the response object for GetDimensions
type GetDimensionsResponse struct {
	cantabular.PaginationResponse
	Dimensions []Dimension `json:"items"`
}
