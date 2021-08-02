package cantabular

// Variable represents a 'variable' object returned from Cantabular
type Variable struct{
	Name         string    `json:"name"`
	Label        string    `json:"label"`
	Len          int       `json:"len"`
	Codes        []string  `json:"codes,omitempty"`
	Labels       []string  `json:"labels,omitempty"`
	MapFrom      []MapFrom `json:"mapFrom,omitempty"`
}

// MapFrom represents the 'mapFrom' object from variable when category
// information is included
type MapFrom struct{
	SourceNames []string `json:"sourceNames,omitempty"`
	Code        []string `json:"codes,omitempty"`
}
