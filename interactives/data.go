package interactives

import "time"

type InteractiveMetadata struct { // TODO : Geography
	Title           string    `bson:"title"                      json:"title"`
	PrimaryTopic    string    `bson:"primary_topic"              json:"primary_topic"`
	Topics          []string  `bson:"topics"                     json:"topics"`
	Surveys         []string  `bson:"surveys"                    json:"surveys"`
	ReleaseDate     time.Time `bson:"release_date"               json:"release_date"`
	Uri             string    `bson:"uri"                        json:"uri"`
	Edition         string    `bson:"edition,omitempty"          json:"edition,omitempty"`
	Keywords        []string  `bson:"keywords,omitempty"         json:"keywords,omitempty"`
	MetaDescription string    `bson:"meta_description,omitempty" json:"meta_description,omitempty"`
	Source          string    `bson:"source,omitempty"           json:"source,omitempty"`
	Summary         string    `bson:"summary,omitempty"          json:"summary,omitempty"`
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
