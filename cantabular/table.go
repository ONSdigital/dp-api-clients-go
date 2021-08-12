package cantabular

// Table represents the 'table' field from the GraphQL dataset
// query response
type Table struct {
	Dimensions []Dimension `json:"dimensions"`
	Values     []int       `json:"values"`
	Error      string      `json:"error,omitempty" `
}
