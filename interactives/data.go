package interactives

type InteractiveUpdated struct {
	ImportStatus bool              `json:"importstatus,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type Interactive struct {
	ID       string            `json:"id,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type List struct {
	Items      []Interactive `json:"items"`
	Count      int           `json:"count"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"total_count"`
}
