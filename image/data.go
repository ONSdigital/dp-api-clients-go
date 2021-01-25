package image

import "time"

// Images represents the fields for a group of images as returned by Image API
type Images struct {
	Count      int     `json:"count"`
	Items      []Image `json:"items"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	TotalCount int     `json:"total_count"`
}

// NewImage represents the fields required to create a new Image
type NewImage struct {
	CollectionId string  `json:"collection_id,omitempty"`
	State        string  `json:"state,omitempty"`
	Filename     string  `json:"filename,omitempty"`
	License      License `json:"license,omitempty"`
	Type         string  `json:"type,omitempty"`
}

// License represents the fields for an Image License
type License struct {
	Title string `json:"title,omitempty"`
	Href  string `json:"href,omitempty"`
}

// Image represents the fields for an Image
type Image struct {
	Id           string `json:"id,omitempty"`
	CollectionId string `json:"collection_id,omitempty"`
	State        string `json:"state,omitempty"`
	//enum:
	//- created
	//- uploaded
	//- importing
	//- imported
	//- published
	//- completed
	//- deleted
	//- failed_import
	//- failed_publish
	Error    string      `json:"error,omitempty"`
	Filename string      `json:"filename,omitempty"`
	License  License     `json:"license,omitempty"`
	Links    *ImageLinks `json:"links,omitempty"`
	Upload   ImageUpload `json:"upload,omitempty"`
	Type     string      `json:"type,omitempty"`
}

// ImageUpload represents the fields for an Image Upload
type ImageUpload struct {
	Path string `json:"path,omitempty"`
}

// ImageLinks represents the fields for the image HATEOAS links
type ImageLinks struct {
	Self      string `json:"self"`
	Downloads string `json:"downloads"`
}

// Images represents the fields for a group of image download variants as returned by Image API
type ImageDownloads struct {
	Count      int             `json:"count"`
	Items      []ImageDownload `json:"items"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	TotalCount int             `json:"total_count"`
}

// ImageDownload represents the fields for an Image Download
type ImageDownload struct {
	Id      string              `json:"id,omitempty"`
	Height  *int                `json:"height,omitempty"`
	Href    string              `json:"href,omitempty"`
	Palette string              `json:"palette,omitempty"`
	Private string              `json:"private,omitempty"`
	Public  bool                `json:"public,omitempty"`
	Size    int                 `json:"size,omitempty"`
	Type    string              `json:"type,omitempty"`
	Width   *int                `json:"width,omitempty"`
	Links   *ImageDownloadLinks `json:"links,omitempty"`
	State   string              `json:"state,omitempty"`
	//enum:
	//- pending
	//- importing
	//- imported
	//- published
	//- completed
	//- failed
	Error            string     `json:"error,omitempty"`
	ImportStarted    *time.Time `json:"import_started,omitempty"`
	ImportCompleted  *time.Time `json:"import_completed,omitempty"`
	PublishStarted   *time.Time `json:"publish_started,omitempty"`
	PublishCompleted *time.Time `json:"publish_completed,omitempty"`
}

// ImageDownload represents the fields for an Image Download
type NewImageDownload struct {
	Id      string `json:"id,omitempty"`
	Height  *int   `json:"height,omitempty"`
	Palette string `json:"palette,omitempty"`
	Size    int    `json:"size,omitempty"`
	Type    string `json:"type,omitempty"`
	Width   *int   `json:"width,omitempty"`
	State   string `json:"state,omitempty"`
	//enum:
	//- importing
	ImportStarted *time.Time `json:"import_started,omitempty"`
}

// ImageDownloadLinks represents the fields for the image download HATEOAS links
type ImageDownloadLinks struct {
	Self  string `json:"self"`
	Image string `json:"image"`
}
