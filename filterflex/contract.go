package filterflex

// AuthHeaders represents the common set of headers required for making
// authorized requests
type AuthHeaders struct {
	UserAuthToken    string
	ServiceAuthToken string
}

// getFilterInput holds the required fields for making the GET /filters
// API call
type GetDeleteOptionInput struct {
	FilterID  string
	Dimension string
	Option    string
	IfMatch   string
	AuthHeaders
}
