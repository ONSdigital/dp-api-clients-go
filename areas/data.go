package areas

// AreaDetails represents a response area model from the areas api
type AreaDetails struct {
	Code          string                 `json:"code,omitempty"`
	Name          string                 `json:"name,omitempty"`
	DateStarted   string                 `json:"date_start,omitempty"`
	DateEnd       string                 `json:"date_end,omitempty"`
	WelshName     string                 `json:"name_welsh,omitempty"`
	GeometricData map[string]interface{} `json:"geometry"`
	Visible       bool                   `json:"visible,omitempty"`
	AreaType      string                 `json:"area_type,omitempty"`
}

// Relation represents a response relation model from area api
type Relation struct {
	AreaCode string `json:"area_code,omitempty"`
	AreaName string `json:"area_name,omitempty"`
	Href     string `json:"href,omitempty"`
}
