package cantabular

// VariableBase represents the minimum amount of fields returned
// by Cantabular for a 'variable' object. Is kept separate from
// the full Variable struct because GraphQL/Cantabular has difficulty
// unmarshaling into a struct with unexpected objects (in this case MapFrom),
// even if they're set to not be included
type VariableBase struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// Variable represents a 'variable' object returned from Cantabular
type Variable struct {
	VariableBase
	Len     int       `json:"len"`
	Codes   []string  `json:"codes,omitempty"`
	Labels  []string  `json:"labels,omitempty"`
	MapFrom []MapFrom `json:"mapFrom,omitempty"`
}

// MapFrom represents the 'mapFrom' object from variable when category
// information is included
type MapFrom struct {
	SourceNames []string `json:"sourceNames,omitempty"`
	Code        []string `json:"codes,omitempty"`
}
