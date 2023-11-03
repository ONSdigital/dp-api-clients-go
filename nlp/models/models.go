package models

type NLPResp struct {
	Berlin   Berlin
	Category Category
	Scrubber Scrubber
}

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
type Category []struct {
	Code  []string `json:"c,omitempty"`
	Score float64  `json:"s,omitempty"`
}

type Scrubber struct {
	Results Results `json:"results,omitempty"`
	Time    string  `json:"time,omitempty"`
	Query   string  `json:"query"`
}

type Results struct {
	Areas      []AreaResp     `json:"areas,omitempty"`
	Industries []IndustryResp `json:"industries,omitempty"`
}

type AreaResp struct {
	Codes      map[string]string `json:"codes,omitempty"`
	Name       string            `json:"name,omitempty"`
	Region     string            `json:"region,omitempty"`
	RegionCode string            `json:"region_code,omitempty"`
}

type IndustryResp struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}
