package gql

type Dataset struct {
	RuleBase  RuleBase  `json:"ruleBase"`
	Variables Variables `json:"variables"`
}

type RuleBase struct {
	IsSourceOf Variables `json:"isSourceOf"`
	Name       string    `json:"name"`
}

type Variables struct {
	Edges          []Edge `json:"edges"`
	Search         Search `json:"search,omitempty"`
	CategorySearch Search `json:"categorySearch,omitempty"`
	TotalCount     int    `json:"totalCount"`
}

type Search struct {
	Edges []Edge `json:"edges"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	Name             string      `json:"name"`
	Code             string      `json:"code"`
	Label            string      `json:"label"`
	Categories       Categories  `json:"categories"`
	MapFrom          []Variables `json:"mapFrom"`
	FilterOnly       string      `json:"filterOnly,omitempty"`
	Variable         Variable    `json:"variable"`
	IsDirectSourceOf Variables   `json:"isDirectSourceOf"`
}

type Categories struct {
	TotalCount int    `json:"totalCount"`
	Edges      []Edge `json:"edges"`
	Search     Search `json:"search,omitempty"`
}

type Variable struct {
	Name string `json:"name"`
}
