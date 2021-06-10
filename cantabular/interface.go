package cantabular

//go:generate moq -out mock/http-client.go . httpClient

import (
	"context"
	"net/http"
)

// httpClient is an interface for a user agent to make http requests
type httpClient interface{
	Get(ctx context.Context, url string) (*http.Response, error)
}

// coder is an interface that allows you to 
// extract a http status code from an error (or other object)
type coder interface{
	Code() int
}

// dataLogger is an interface that allows you to 
// extract logData from an error (or other object)
type dataLogger interface{
	LogData() map[string]interface{}
}