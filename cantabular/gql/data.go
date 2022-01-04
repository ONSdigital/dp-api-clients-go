package gql

type Dataset struct {
	RuleBase RuleBase `json:"ruleBase"`
}

type RuleBase struct {
	IsSourceOf MapFrom `json:"isSourceOf"`
	Name       string  `json:"name"`
}

type MapFrom struct {
	Edges []Edge `json:"edges"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	Name       string     `json:"name"`
	Label      string     `json:"label"`
	Categories Categories `json:"categories"`
	MapFrom    []MapFrom  `json:"mapFrom"`
	FilterOnly string     `json:"filterOnly"`
}

type Categories struct {
	TotalCount int `json:"totalCount"`
}
