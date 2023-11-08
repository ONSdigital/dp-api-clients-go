package models

type Category struct {
	Code  []string `json:"c,omitempty"`
	Score float64  `json:"s,omitempty"`
}
