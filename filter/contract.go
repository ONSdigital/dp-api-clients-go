package filter

// Find in this file up to date request/responses to some of the dp-filter-api
// endpoints. For some reason the existing functions are largely out of date
// with how the API actually behaves, but are going to create fresh functions
// for interacting rather than update old ones in case something somewhere breaks.
// Perhaps the old ones can be removed at a later date once we're sure they're
// not needed.

// AuthHeaders represents the common set of headers required for making
// authorized requests
type AuthHeaders struct {
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
}

// getFilterInput holds the required fields for making the GET /filters
// API call
type GetFilterInput struct {
	FilterID string
	AuthHeaders
}

// getFilterResponse is the response body for GET /filters
type GetFilterResponse struct {
	ID             string      `json:"id"`
	FilterID       string      `json:"filter_id"`
	InstanceID     string      `json:"instance_id"`
	Links          FilterLinks `json:"links"`
	Dataset        Dataset     `json:"dataset,omitempty"`
	State          string      `json:"state"`
	Published      bool        `json:"published"`
	PopulationType string      `json:"population_type,omitempty"`
	Custom         *bool       `json:"custom,omitempty"`
	ETag           string      `json:"-"`
}

// FilterLinks represents the links object for /filters related endpoints
type FilterLinks struct {
	Dimensions Link `json:"dimensions,omitempty"`
	Self       Link `json:"self,omitempty"`
	Version    Link `json:"version,omitempty"`
}

type CreateFlexBlueprintRequest struct {
	Dataset        Dataset          `json:"dataset"`
	Dimensions     []ModelDimension `json:"dimensions"`
	PopulationType string           `json:"population_type"`
	Custom         bool             `json:"custom"`
	CollectionID   string           `json:"-"`
}

type createFlexBlueprintResponse struct {
	FilterID string `json:"filter_id"`
}

// createFlexDimensionRequest represents the fields required to add a dimension to a flex filter
type createFlexDimensionRequest struct {
	Name       string   `json:"name"`
	IsAreaType bool     `json:"is_area_type"`
	Options    []string `json:"options"`
}
