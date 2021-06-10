package recipe

// ErrorResponse models the error response from cantabular
type ErrorResponse struct {
	Message string `json:"message"`
}

// Recipe holds the response body for GET /recipes/{id}
type Recipe struct {
	ID              string     `json:"id,omitempty"`
	Alias           string     `json:"alias,omitempty"`
	Format          string     `json:"format,omitempty"`
	InputFiles      []file     `json:"files,omitempty"`
	OutputInstances []Instance `json:"output_instances,omitempty"`
	CantabularBlob  string     `json:"cantabular_blob,omitempty"`
}

// CodeList holds one of the codelists corresponding to a recipe
type CodeList struct {
	ID          string `json:"id,omitempty"`
	HRef        string `json:"href,omitempty"`
	Name        string `json:"name,omitempty"`
	IsHierarchy *bool  `json:"is_hierarchy,omitempty"`
}

// Instance holds one of the output_instances corresponding to a recipe
type Instance struct {
	DatasetID string     `json:"dataset_id,omitempty"`
	Editions  []string   `json:"editions,omitempty"`
	Title     string     `json:"title,omitempty"`
	CodeLists []CodeList `json:"code_lists,omitempty"`
}

// file holds one of the file descriptions corresponding to a recipe
type file struct {
	Description string `json:"description,omitempty"`
}
