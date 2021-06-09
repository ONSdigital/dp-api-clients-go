package cantabular

import (
	"time"
)

// Dataset represents a 'dataset' object returned from Cantabular
type Dataset struct{
	Name              string    `json:"name"`
	Digest            string    `json:"digest"`
	Description       string    `json:"description"`
	Size              int       `json:"size"`
	RulebaseVariable  string    `json:"ruleBaseVariable"`
	DateTime          time.Time `json:"datetime"`
}
