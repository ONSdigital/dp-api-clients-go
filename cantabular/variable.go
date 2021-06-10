package cantabular

// Variable represents a 'variable' object returned from Cantabular
type Variable struct{
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Len          int      `json:"len"`
	Codes        []string `json:"codes,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	MapFrom      []string `json:"mapFrom,omitempty"`
	MapFromCodes []string `json:"mapFromCodes,omitempty"`
}
