package areas

// AreaDetails represents a response area model from the areas api
type AreaDetails struct {
	Code          string                              `json:"code,omitempty"`
	Name          string                              `json:"name,omitempty"`
	DateStarted   string                              `json:"date_start,omitempty"`
	DateEnd       string                              `json:"date_end,omitempty"`
	WelshName     string                              `json:"name_welsh,omitempty"`
	GeometricData []map[string]map[string]interface{} `json:"features"`
	Visible       bool                                `json:"visible,omitempty"`
	AreaType      string                              `json:"area_type",omitempty"`
}
