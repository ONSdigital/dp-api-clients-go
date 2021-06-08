package recipe

//Recipe - struct for individual recipe
type Recipe struct {
	ID              string     `json:"id,omitempty"`
	Alias           string     `json:"alias,omitempty"`
	Format          string     `json:"format,omitempty"`
	InputFiles      []file     `json:"files,omitempty"`
	OutputInstances []Instance `json:"output_instances,omitempty"`
	CantabularBlob  string     `json:"cantabular_blob,omitempty"`
}

//CodeList - Code lists for instance
type CodeList struct {
	ID          string `json:"id,omitempty"`
	HRef        string `json:"href,omitempty"`
	Name        string `json:"name,omitempty"`
	IsHierarchy *bool  `json:"is_hierarchy,omitempty"`
}

//Instance - struct for instance of recipe
type Instance struct {
	DatasetID string     `json:"dataset_id,omitempty"`
	Editions  []string   `json:"editions,omitempty"`
	Title     string     `json:"title,omitempty"`
	CodeLists []CodeList `json:"code_lists,omitempty"`
}

type file struct {
	Description string `json:"description,omitempty"`
}
