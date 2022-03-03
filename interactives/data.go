package interactives

import "time"

type InteractiveMetadata struct { // TODO : Geography
	Title           string    `json:"title"`
	PrimaryTopic    string    `json:"primary_topic"`
	Topics          []string  `json:"topics"`
	Surveys         []string  `json:"surveys"`
	ReleaseDate     time.Time `json:"release_date"`
	Uri             string    `json:"uri"`
	Edition         string    `json:"edition,omitempty"`
	Keywords        []string  `json:"keywords,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	Source          string    `json:"source,omitempty"`
	Summary         string    `json:"summary,omitempty"`
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
	Interactive      Interactive `json:"interactive,omitempty"`
}

type List struct {
	Items      []Interactive `json:"items"`
	Count      int           `json:"count"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"total_count"`
}
