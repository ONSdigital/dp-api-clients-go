package dimension

import "strings"

// ErrorResp represents an error response containing a list of errors
type ErrorResp struct {
	Errors []string `json:"errors"`
}

func (e ErrorResp) Error() string {
	return strings.Join(e.Errors, ", ")
}
