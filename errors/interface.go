package errors

// coder is an interface that allows you to 
// extract a http status code from an error (or other object)
// TODO: Would preferably be called statusCoder defining StatusCode()
// but Code() is more common with ONS code at the moment and should
// be part of a broader change if implemented
type coder interface{
	Code() int
}

// dataLogger is an interface that allows you to 
// extract logData from an error (or other object)
type dataLogger interface{
	LogData() map[string]interface{}
}