package cantabular

// Table represents the 'table' field from the GraphQL dataset
// query response
type Table struct {
	Dimensions []Dimension
	Values     []int
	Error      string
}
