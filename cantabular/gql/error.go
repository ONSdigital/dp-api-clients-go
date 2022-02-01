package gql

import (
	"net/http"
	"strconv"
)

type Error struct {
	Message   string     `json:"message"`
	Locations []Location `json:"locations"`
	Path      []string   `json:"path"`
}

// StatusCode returns the status code defined at the begining of the Error message.
// For example: a status 404 is extracted from '404 Not Found: dataset not loaded in this server'.
// If no status code is provided, then a value of 502 bad gateway is returned.
func (e *Error) StatusCode() int {
	if len(e.Message) < 3 {
		return http.StatusBadGateway
	}

	statusCode, err := strconv.Atoi(e.Message[:3])
	if err != nil {
		return http.StatusBadGateway
	}

	if http.StatusText(statusCode) == "" {
		return http.StatusBadGateway
	}

	return statusCode
}

// 404 Not Found: dataset not loaded in this server

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
