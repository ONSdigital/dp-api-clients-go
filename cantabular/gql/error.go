package gql

type Error struct {
	Message   string     `json:"message"`
	Locations []Location `json:"locations"`
	Path      []string   `json:"path"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
