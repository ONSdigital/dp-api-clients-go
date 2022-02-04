package filterflex

// Error is the package's error type
type Error struct {
	err        error
	logData    map[string]interface{}
}

// Error implements the standard Go error
func (e *Error) Error() string {
	if e.err == nil{
		return "nil"
	}
	return e.err.Error()
}

// Unwrap implements Go error unwrapping
func (e *Error) Unwrap() error {
	return e.err
}

// LogData implemented the DataLogger interface and allows
// log data to be embedded in and retrieved from an error
func (e *Error) LogData() map[string]interface{} {
	return e.logData
}
