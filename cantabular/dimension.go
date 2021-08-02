package cantabular

// Dimension represents the 'dimension' field from a GraphQL
// query dataset response
type Dimension struct{
	Count      int        `json:"count"`
	Categories []Category `json:"categories"`
	Variable   Variable   `json:"variable"`
}
