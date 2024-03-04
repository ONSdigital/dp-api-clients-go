package cantabularmetadata

import (
	"context"
	"io"
	"net/http"
)

// httpClient is an interface for a user agent to make http requests
type httpClient interface {
	Get(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
}

type coder interface {
	Code() int
}
