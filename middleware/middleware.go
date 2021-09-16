package middleware

import (
	"net/http"
)

// Allowed provides a list of methods for which the handler should be executed
type Allowed struct {
	Methods []string
	Handler func(w http.ResponseWriter, req *http.Request)
}

// isMethodAllowed determines if a method is allowed or not
func (a *Allowed) isMethodAllowed(method string) bool {
	for _, s := range a.Methods {
		if method == s {
			return true
		}
	}
	return false
}

// HealthcheckFilter creates a map for healthcheck endpoints whitelisting, to be provided to the Whitelist generic function
func HealthcheckFilter(hcHandler func(w http.ResponseWriter, req *http.Request)) map[string]Allowed {
	return map[string]Allowed{
		"/health": {
			Methods: []string{http.MethodGet},
			Handler: hcHandler,
		},
	}
}

// Whitelist creates a middleware that executes whitelisted endpoints
// The provided whitelist is keyed by path, and contains the handler to use and the methods for which the whitelist applies
func Whitelist(whitelist map[string]Allowed) func(h http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			if wh, ok := whitelist[req.URL.Path]; ok && wh.isMethodAllowed(req.Method) {
				wh.Handler(w, req)
				return
			}

			nextHandler.ServeHTTP(w, req)
		})
	}
}
