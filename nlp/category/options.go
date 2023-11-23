package category

import (
	"net/http"
	"net/url"
)

// Options is a struct containing for customised options for the API client
type Options struct {
	Headers http.Header
	Query   url.Values
}

// Q sets the 'q' Query parameter to the request
func (o *Options) Q(val string) *Options {
	o.Query.Set("query", val)
	return o
}

func setHeaders(req *http.Request, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
}
