package recipe

import (
	"errors"
	"net/http"
)

// Error is the package's error type
type Error struct {
	err        error
	statusCode int
	logData    map[string]interface{}
}

// Error implements the standard Go error
func (e *Error) Error() string {
	return e.err.Error()
}

// Unwrap implements Go error unwrapping
func (e *Error) Unwrap() error {
	return e.err
}

// Code returns the statusCode returned by the Recipe API.
func (e *Error) Code() int {
	return e.statusCode
}

// LogData implemented the DataLogger interface and allows
// log data to be embedded in and retrieved from an error
func (e *Error) LogData() map[string]interface{} {
	return e.logData
}

// coder is an interface that allows you to
// extract a http status code from an error (or other object)
type coder interface {
	Code() int
}

// StatusCode is a callback function that allows you to extract
// a status code from an error, or returns 500 as a default
func StatusCode(err error) int {
	var cerr coder
	if errors.As(err, &cerr) {
		return cerr.Code()
	}

	return http.StatusInternalServerError
}
