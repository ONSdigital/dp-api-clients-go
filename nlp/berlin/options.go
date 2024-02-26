package berlin

import (
	"net/http"
	"net/url"
)

// Options is a struct containing for customised options for the API client
type Options struct {
	Headers http.Header
	Query   url.Values
}

// empty Options
func OptInit() Options {
	return Options{
		Query:   url.Values{},
		Headers: http.Header{},
	}
}

// Q sets the 'q' Query parameter to the request
// Required
func (o *Options) Q(val string) *Options {
	o.Query.Set("q", val)
	return o
}

// State sets the 'state' Query parameter to the request
// Optional default is 'gb'
func (o *Options) State(val string) *Options {
	o.Query.Set("state", val)
	return o
}

// LevDist sets the 'lev_distance' Query parameter to the request
// Optional default is '2'
func (o *Options) LevDist(val string) *Options {
	o.Query.Set("lev_distance", val)
	return o
}

// Limit sets the 'limit' Query parameter to the request
// Optional default is '10'
func (o *Options) Limit(val string) *Options {
	o.Query.Set("limit", val)
	return o
}

func setHeaders(req *http.Request, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
}
