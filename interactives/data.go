package interactives

const (
	UpdateFormFieldKey = "interactive"
	PatchArchive       = "Archive"
)

type PatchRequest struct {
	Attribute   string      `json:"attribute,omitempty"`
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
	Name             string             `json:"name,omitempty"`
	Size             int64              `json:"size_in_bytes,omitempty"`
	Files            []*InteractiveFile `json:"files,omitempty"`
	ImportMessage    string             `json:"import_message,omitempty"`
	ImportSuccessful bool               `json:"import_successful,omitempty"`
}

type InteractiveFile struct {
	Name     string `json:"name,omitempty"`
	Mimetype string `json:"mimetype,omitempty"`
	Size     int64  `json:"size_in_bytes,omitempty"`
	URI      string `json:"uri,omitempty"`
}
