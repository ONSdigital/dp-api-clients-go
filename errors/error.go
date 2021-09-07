package errors

import (
	"errors"
)

// Error is a common error for HTTP APIs to return associated status code
// and logdata from API calls made
type Error struct {
	err        error
	statusCode int
	logData    map[string]interface{}
}

// New a new Error
func New(err error, statusCode int, logData map[string]interface{}) *Error {
	if err == nil {
		err = errors.New("nil error")
	}
	return &Error{
		err:        err,
		statusCode: statusCode,
		logData:    logData,
	}
}

// Error implements the standard Go error
func (e *Error) Error() string {
	return e.err.Error()
}

// Unwrap implements Go error unwrapping
func (e *Error) Unwrap() error {
	return e.err
}

// Code returns the statusCode returned by Cantabular.
// Hopefull can be renamed to StatusCodea some point but this is
// how it is named elsewhere across ONS services and is more useful
// being consistent
func (e *Error) Code() int {
	return e.statusCode
}

// LogData implemented the DataLogger interface and allows
// log data to be embedded in and retrieved from an error
func (e *Error) LogData() map[string]interface{} {
	return e.logData
}
