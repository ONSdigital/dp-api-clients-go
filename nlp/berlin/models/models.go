package models

type Berlin struct {
	Matches []Matches `json:"matches,omitempty"`
	Query   string    `json:"query,omitempty"`
}

type Matches struct {
	Loc    Locations `json:"loc,omitempty"`
	Scores Scores    `json:"scores,omitempty"`
}

type Locations struct {
	Codes       []string `json:"codes,omitempty"`
	Encoding    string   `json:"encoding,omitempty"`
	Names       []string `json:"names,omitempty"`
	ID          string   `json:"id,omitempty"`
	Key         string   `json:"key,omitempty"`
	State       []string `json:"state,omitempty"`
	Subdivision []string `json:"subdiv,omitempty"`
	Words       []string `json:"words,omitempty"`
}

type Scores struct {
	Offset []int `json:"offset,omitempty"`
	Score  int   `json:"score,omitempty"`
}
