package image

type Images struct {
	Count      int     `json:"count"`
	Items      []Image `json:"items"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	TotalCount int     `json:"total_count"`
}

type NewImage struct {
	CollectionId string  `json:"collection_id,omitempty"`
	Filename string  `json:"filename,omitempty"`
	License License   `json:"license,omitempty"`
	Type string  `json:"type,omitempty"`
}

type License struct {
	Title string  `json:"title,omitempty"`
	Href string  `json:"href,omitempty"`
}

type Image struct {
	Id string  `json:"id,omitempty"`
	CollectionId	 string  `json:"collection_id,omitempty"`
	State string  `json:"state,omitempty"`
	//enum:
	//- created
	//- uploaded
	//- publishing
	//- published
	//- deleted
	Filename string  `json:"filename,omitempty"`
	License License   `json:"license,omitempty"`
	Upload ImageUpload `json:"upload,omitempty"`
	Type string  `json:"type,omitempty"`
	Downloads ImageDownloads  `json:"downloads,omitempty"`
}

type ImageUpload struct {
	Path string  `json:"path,omitempty"`
}

type ImageDownloads struct {
	Png ImagedDownloadsVariant  `json:"png,omitempty"`
}

type ImagedDownloadsVariant struct {
	Size1920x1080 ImageDownload  `json:"1920x1080,omitempty"`
	Size1280x720 ImageDownload  `json:" 1280x720,omitempty"`
	ThumbNail ImageDownload  `json:"thumbnail,omitempty"`
}

type ImageDownload struct {
	Size int `json:"size,omitempty"`
	Href string `json:"href,omitempty"`
	Public string `json:"public,omitempty"`
	Private string `json:"private,omitempty"`
}