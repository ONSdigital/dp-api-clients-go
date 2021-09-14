package cantabular

// Category represents the 'category' field from the GraphQL
// query dataset response
type Category struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}
