package errors

import (
	navtiveErrors "errors"
	"fmt"
	"strings"
)

type JsonError struct {
	Code        string `json:"errorCode"`
	Description string `json:"description"`
}

type JsonErrors struct {
	Errors []JsonError `json:"errors"`
}

func (j JsonErrors) ToNativeError() error {
	var msgs []string
	for _, e := range j.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Description))
	}
	return navtiveErrors.New(strings.Join(msgs, "\n"))
}
