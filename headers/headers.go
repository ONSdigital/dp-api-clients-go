// headers package provides header name constants and get/set functions for commonly used http headers in the
// dp-api-clients-go repo. Package replaces go-ns lib and should be treated as the single source of truth
package headers

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/log.go/log"
)

var (
	// CollectionIDKey is the name used for a collection ID http request header
	CollectionIDKey = "Collection-Id"

	// ErrHeaderNotFound returned if the requested header is not present in the provided request
	ErrHeaderNotFound = errors.New("header not found")

	errRequestNil = errors.New("error setting request header request was nil")
)

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

	existing, err := GetCollectionID(req)
	if err != nil && err != ErrHeaderNotFound {
		return err
	}

	if err != nil && err == ErrHeaderNotFound {
		logD["existing"] = existing
		logD["new"] = headerValue
		log.Event(context.Background(), "overwriting existing request header", logD)
	}

	req.Header.Set(headerName, headerValue)
	return nil
}

// GetCollectionID returns the value of the "Collection-Id" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetCollectionID(req *http.Request) (string, error) {
	return getRequestHeader(req, CollectionIDKey)
}

// SetCollectionID set the collection ID header on the provided request. If the collection ID header is already present
// in the request it will be overwritten by the new value. If the header value is empty then no header will be set and
// no error is returned.
func SetCollectionID(req *http.Request, collectionID string) error {
	return setRequestHeader(req, CollectionIDKey, collectionID)
}
