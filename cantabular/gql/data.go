package gql

type DatasetRuleBase struct {
	RuleBase RuleBase `json:"ruleBase"`
}

type DatasetVariables struct {
	Variables Variables `json:"variables"`
}

type RuleBase struct {
	IsSourceOf Variables `json:"isSourceOf"`
	Name       string    `json:"name"`
}

type Variables struct {
	Edges []Edge `json:"edges"`
}

type Edge struct {
	Node Node `json:"node"`
}

type Node struct {
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Categories Categories  `json:"categories"`
	MapFrom    []Variables `json:"mapFrom"`
	FilterOnly string      `json:"filterOnly"`
}

type Categories struct {
	TotalCount int `json:"totalCount"`
}
