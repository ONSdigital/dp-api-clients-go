// headers package provides header name constants and get/set functions for commonly used http headers in the
// dp-api-clients-go repo. Package replaces go-ns lib and should be treated as the single source of truth
package headers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/ONSdigital/log.go/log"
)

const (
	// collectionID is the name used for a collection ID http request header
	collectionID = "Collection-Id"

	// userAuthToken is the user Florence auth token header name
	userAuthToken = "X-Florence-Token"

	// serviceAuthToken the service auth token header name
	serviceAuthToken = "Authorization"

	// bearerPrefix is the prefix for authorization header values
	bearerPrefix = "Bearer "
)

var (
	// ErrHeaderNotFound returned if the requested header is not present in the provided request
	ErrHeaderNotFound = errors.New("header not found")

	errRequestNil = errors.New("error setting request header request was nil")
)

// GetCollectionID returns the value of the "Collection-Id" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetCollectionID(req *http.Request) (string, error) {
	return getRequestHeader(req, collectionID)
}

// SetCollectionID set the collection ID header on the provided request. If the collection ID header is already present
// in the request it will be overwritten by the new value. If the header value is empty then no header will be set and
// no error is returned.
func SetCollectionID(req *http.Request, headerValue string) error {
	return setRequestHeader(req, collectionID, headerValue)
}

// SetUserAuthToken set the user authentication token header on the provided request. If the authentication token is
// already present it will be overwritten by the new value. If the header value is empty then no header will be set and
// no error is returned.
func SetUserAuthToken(req *http.Request, headerValue string) error {
	return setRequestHeader(req, userAuthToken, headerValue)
}

// GetUserAuthToken returns the value of the "X-Florence-Token" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetUserAuthToken(req *http.Request) (string, error) {
	return getRequestHeader(req, userAuthToken)
}

// SetServiceAuthToken set the service authentication token header on the provided request. If the authentication token is
// already present it will be overwritten by the new value. If the header value is empty then no header will be set and
// no error is returned.
func SetServiceAuthToken(req *http.Request, headerValue string) error {
	if req == nil {
		return errRequestNil
	}

	if len(headerValue) == 0 {
		log.Event(context.Background(), "request header not set as value was empty", log.Data{
			"header_name": serviceAuthToken,
		})
		return nil
	}

	if !strings.HasPrefix(headerValue, bearerPrefix) {
		headerValue = bearerPrefix + headerValue
	}

	return setRequestHeader(req, serviceAuthToken, headerValue)
}

// GetServiceAuthToken returns the value of the "Authorization" request header if it exists, returns ErrHeaderNotFound if
// the header is not found. If the header exists the "Bearer " prefixed is removed from returned value.
func GetServiceAuthToken(req *http.Request) (string, error) {
	token, err := getRequestHeader(req, serviceAuthToken)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(token, bearerPrefix) {
		token = strings.TrimPrefix(token, bearerPrefix)
	}
	return token, nil
}

func getRequestHeader(req *http.Request, headerName string) (string, error) {
	if req == nil {
		return "", errRequestNil
	}

	headerValue := req.Header.Get(headerName)
	if len(headerValue) == 0 {
		return "", ErrHeaderNotFound
	}

	return headerValue, nil
}

func setRequestHeader(req *http.Request, headerName string, headerValue string) error {
	if req == nil {
		return errRequestNil
	}

	logD := log.Data{"header_name": headerName}

	if len(headerValue) == 0 {
		log.Event(context.Background(), "request header not set as value was empty", logD)
		return nil
	}

	existing := req.Header.Get(headerName)
	if len(existing) > 0 {
		logD["existing"] = existing
		logD["new"] = headerValue
		log.Event(context.Background(), "overwriting existing request header", logD)
	}

	req.Header.Set(headerName, headerValue)
	return nil
}
