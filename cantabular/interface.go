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
