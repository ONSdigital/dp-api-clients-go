package gql

type Dataset struct {
	Name        string    `json:"name,omitempty"`
	Label       string    `json:"label,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type,omitempty"`
	RuleBase    RuleBase  `json:"ruleBase"`
	Variables   Variables `json:"variables"`
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

type Meta struct {
	ONSVariable ONS_Variable `json:"ONS_Variable"`
}

type ONS_Variable struct {
	GeographyHierarchyOrder string `json:"Geography_Hierarchy_Order"`
	QualityStatementText    string `json:"quality_statement_text"`
}

type Node struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Code             string      `json:"code"`
	Label            string      `json:"label"`
	Categories       Categories  `json:"categories"`
	MapFrom          []Variables `json:"mapFrom"`
	Variable         Variable    `json:"variable"`
	IsDirectSourceOf Variables   `json:"isDirectSourceOf"`
	IsSourceOf       Variables   `json:"isSourceOf"`
	Meta             Meta        `json:"meta"`
}

type Categories struct {
	TotalCount int    `json:"totalCount"`
	Edges      []Edge `json:"edges"`
	Search     Search `json:"search,omitempty"`
}

type Variable struct {
	Name string `json:"name"`
}
