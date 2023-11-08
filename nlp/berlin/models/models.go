package models

type Berlin struct {
	Matches []Matches `json:"matches,omitempty"`
}

type Matches struct {
	Codes       []string `json:"codes,omitempty"`
	Encoding    string   `json:"encoding,omitempty"`
	Names       []string `json:"names,omitempty"`
	ID          string   `json:"id,omitempty"`
	Key         string   `json:"key,omitempty"`
	State       []string `json:"state,omitempty"`
	Subdivision []string `json:"subdiv,omitempty"`
	Words       []string `json:"words,omitempty"`
}
