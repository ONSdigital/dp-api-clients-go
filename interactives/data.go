package interactives

const (
	UpdateFormFieldKey = "interactive"

	PatchArchive     PatchAttribute = "Archive"
	Publish          PatchAttribute = "Publish"
	LinkToCollection PatchAttribute = "LinkToCollection"
)

type PatchAttribute string

type PatchRequest struct {
	Attribute    PatchAttribute `json:"attribute,omitempty"`
	Interactive  Interactive    `json:"interactive,omitempty"`
}

type Filter struct {
	AssociateCollection bool      `json:"associate_collection,omitempty"`
	Metadata            *Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	Title             string `json:"title,omitempty"`
	Label             string `json:"label,omitempty"`
	InternalID        string `json:"internal_id,omitempty"`
	CollectionID      string `json:"collection_id,omitempty"`
	HumanReadableSlug string `json:"slug,omitempty"`
	ResourceID        string `json:"resource_id,omitempty"`
}

type Interactive struct {
	ID        string      `json:"id,omitempty"`
	Published *bool       `json:"published,omitempty"`
	Metadata  *Metadata   `json:"metadata,omitempty"`
	Archive   *Archive    `json:"archive,omitempty"`
	HTMLFiles []*HTMLFile `json:"html_files,omitempty"`
}

type Archive struct {
	Name                string `json:"name,omitempty"`
	Size                int64  `json:"size_in_bytes,omitempty"`
	UploadRootDirectory string `json:"upload_root_directory,omitempty"`
	ImportMessage       string `json:"import_message,omitempty"`
	ImportSuccessful    bool   `json:"import_successful,omitempty"`
}

type HTMLFile struct {
	Name string `json:"name,omitempty"`
	URI  string `json:"uri,omitempty"`
}
