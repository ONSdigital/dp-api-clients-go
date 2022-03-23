package interactives

const (
	UpdateFormFieldKey = "update"
)

type InteractiveMetadata struct {
	Label      string `json:"label"`
	InternalID string `json:"internal_id"`
}

type Interactive struct {
	ID       string               `json:"id,omitempty"`
	Metadata *InteractiveMetadata `json:"metadata,omitempty"`
	Archive  *InteractiveArchive  `json:"archive,omitempty"`
}

type InteractiveArchive struct {
	Name  string             `json:"name,omitempty"`
	Size  int64              `json:"size_in_bytes,omitempty"`
	Files []*InteractiveFile `json:"files,omitempty"`
}

type InteractiveFile struct {
	Name     string `json:"name,omitempty"`
	Mimetype string `json:"mimetype,omitempty"`
	Size     int64  `json:"size_in_bytes,omitempty"`
}

type InteractiveUpdate struct {
	ImportSuccessful *bool       `json:"import_successful,omitempty"`
	ImportMessage    string      `json:"import_message,omitempty"`
	Interactive      Interactive `json:"interactive,omitempty"`
}

type List struct {
	Items      []Interactive `json:"items"`
	Count      int           `json:"count"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"total_count"`
}
