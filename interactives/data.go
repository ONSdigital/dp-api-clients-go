package interactives

const (
	UpdateFormFieldKey = "interactive"
	PatchImportArchive = "ImportArchive"
)

type PatchRequest struct {
	Action      string      `json:"action,omitempty"`
	Successful  bool        `json:"successful,omitempty"`
	Message     string      `json:"message,omitempty"`
	Interactive Interactive `json:"interactive,omitempty"`
}

type InteractiveFilter struct {
	AssociateCollection bool                 `json:"associate_collection,omitempty"`
	Metadata            *InteractiveMetadata `json:"metadata,omitempty"`
}

type InteractiveMetadata struct {
	Title             string `json:"title,omitempty"`
	Label             string `json:"label,omitempty"`
	InternalID        string `json:"internal_id,omitempty"`
	CollectionID      string `json:"collection_id,omitempty"`
	HumanReadableSlug string `json:"slug,omitempty"`
	ResourceID        string `json:"resource_id,omitempty"`
}

type Interactive struct {
	ID        string               `json:"id,omitempty"`
	Published *bool                `json:"published,omitempty"`
	Metadata  *InteractiveMetadata `json:"metadata,omitempty"`
	Archive   *InteractiveArchive  `json:"archive,omitempty"`
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

type List struct {
	Items      []Interactive `json:"items"`
	Count      int           `json:"count"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"total_count"`
}
