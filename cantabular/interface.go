package cantabular

//go:generate moq -out graphql_client_mock_test.go . GraphQLClient

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

// GraphQLClient is the Client used by the GraphQL package to make queries
type GraphQLClient interface {
	Query(ctx context.Context, query interface{}, vars map[string]interface{}) error
}

type coder interface {
	Code() int
}
