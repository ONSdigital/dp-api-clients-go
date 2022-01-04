package areas

// AreaDetails represents a response area model from the areas api
type AreaDetails struct {
	Code        string `json:"code,omitempty"`
	Name        string `json:"name,omitempty"`
	DateStarted string `json:"date_started,omitempty"`
	DateEnd     string `json:"date_end,omitempty"`
	NameWelsh   string `json:"name_welsh,omitempty"`
	Limit       string `json:"limit,omitempty"`
	TotalCount  string `json:"total_count,omitempty"`
	Visible     bool   `json:"visible,omitempty"`
}
