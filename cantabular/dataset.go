package cantabular

import (
	"time"
)

// Dataset represents a 'dataset' object returned from Cantabular Server
type Dataset struct{
	Name              string    `json:"name"`
	Digest            string    `json:"digest"`
	Description       string    `json:"description"`
	Size              int       `json:"size"`
	RulebaseVariable  string    `json:"ruleBaseVariable"`
	DateTime          time.Time `json:"datetime"`
}

// GQLDataset represents the 'dataset' field from a GraphQL static dataset
// query response
type StaticDataset struct{
	Table Table `json:"table"`
}
