package cantabular

// Error is the package's error type
type Error struct{
	err error
	statusCode int
	logData map[string]interface{}
}

// Error implements the standard Go error
func (e *Error) Error() string {
	return e.err.Error()
}

// Unwrap implements Go error unwrapping
func (e *Error) Unwrap() error{
	return e.err
}

// Code returns the statusCode returned by Cantabular.
// Hopefull can be renamed to StatusCodea some point but this is
// how it is named elsewhere across ONS services and is more useful
// being consistent
func (e *Error) Code() int{
	return e.statusCode
}

// LogData implemented the DataLogger interface and allows
// log data to be embedded in and retrieved from an error
func (e *Error) LogData() map[string]interface{}{
	return e.logData
}