package files

type FileMetaData struct {
	Path          string  `json:"path"`
	IsPublishable bool    `json:"is_publishable"`
	CollectionID  *string `json:"collection_id,omitempty"`
	Title         string  `json:"title"`
	SizeInBytes   uint64  `json:"size_in_bytes"`
	Type          string  `json:"type"`
	Licence       string  `json:"licence"`
	LicenceUrl    string  `json:"licence_url"`
	State         string  `json:"state"`
	Etag          string  `json:"etag"`
}
