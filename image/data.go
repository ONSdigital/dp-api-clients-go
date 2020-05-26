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
	Filename     string  `json:"filename,omitempty"`
	License      License `json:"license,omitempty"`
	Type         string  `json:"type,omitempty"`
}

type License struct {
	Title string `json:"title,omitempty"`
	Href  string `json:"href,omitempty"`
}

type Image struct {
	Id           string `json:"id,omitempty"`
	CollectionId string `json:"collection_id,omitempty"`
	State        string `json:"state,omitempty"`
	//enum:
	//- created
	//- uploaded
	//- publishing
	//- published
	//- deleted
	Filename  string                              `json:"filename,omitempty"`
	License   License                             `json:"license,omitempty"`
	Upload    ImageUpload                         `json:"upload,omitempty"`
	Type      string                              `json:"type,omitempty"`
	Downloads map[string]map[string]ImageDownload `json:"downloads,omitempty"`
}

type ImageUpload struct {
	Path string `json:"path,omitempty"`
}

type ImageDownload struct {
	Size    int    `json:"size,omitempty"`
	Href    string `json:"href,omitempty"`
	Public  string `json:"public,omitempty"`
	Private string `json:"private,omitempty"`
}
