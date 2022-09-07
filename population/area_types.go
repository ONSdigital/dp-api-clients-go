package population

import "github.com/ONSdigital/dp-api-clients-go/v2/cantabular"

// AreaType is an area type model with ID and Label
type AreaType struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TotalCount int    `json:"total_count"`
}

// GetAreaTypesResponse is the response object for GET /area-types
type GetAreaTypesResponse struct {
	cantabular.PaginationResponse
	AreaTypes []AreaType `json:"area_types"`
}
