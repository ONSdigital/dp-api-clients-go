// Package headers package provides header name constants and get/set functions for commonly used http headers in the
// dp-api-clients-go repo. Package replaces go-ns lib and should be treated as the single source of truth
package headers

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// collectionIDHeader is the name used for a collection ID http request header
	collectionIDHeader = "Collection-Id"

	// userAuthTokenHeader is the user Florence auth token header name
	userAuthTokenHeader = "X-Florence-Token"

	// serviceAuthToken the service auth token header name
	serviceAuthTokenHeader = "Authorization"

	// Same value as serviceAuthTokenHeader as this is a standard header but not linked to service authentication
	authTokenHeader = "Authorization"

	// idTokenHeader the unique ID of the signed in user
	idTokenHeader = "ID"

	// refreshTokenHeader a long life token used to get new session tokens
	refreshTokenHeader = "Refresh"

	// bearerPrefix is the prefix for authorization header values
	bearerPrefix = "Bearer "

	// downloadServiceToken is the authorization header for the download service
	downloadServiceTokenHeader = "X-Download-Service-Token"

	// userIdentity is the user identity header used to forward a confirmed identity to another API
	userIdentityHeader = "User-Identity"

	// requestIDHeader is the unique request ID header name
	requestIDHeader = "X-Request-Id"

	// localeCodeHeader is the locale code header name
	localeCodeHeader = "LocaleCode"

	// acceptedLangHeader indicates the natural language and locale that the client prefers
	acceptedLangHeader = "Accept-Language"

	// ifMatchHeader is the If-Match header name
	ifMatchHeader = "If-Match"

	// eTagHeader is the ETag header name
	eTagHeader = "ETag"

	// eTagHeader is the Accept header name
	acceptHeader = "Accept"
)

const (
	// IfMatchAnyETag is a wildchar value for If-Match header to ask the API to ignore the ETag check
	IfMatchAnyETag = "*"
)

var (
	// ErrHeaderNotFound returned if the requested header is not present in the provided request
	ErrHeaderNotFound = errors.New("header not found")

	// ErrValueEmpty returned if an empty value is passed when a non-empty value is required
	ErrValueEmpty = errors.New("header not set as value was empty")

	// ErrRequestNil return if SetX header function is called with a nil request
	ErrRequestNil = errors.New("error setting request header request was nil")

	// ErrResponseNil return if GetResponseX header function is called with a nil response
	ErrResponseNil = errors.New("error getting request header, response was nil")
)

// IsErrNotFound return true if the err equal to ErrHeaderNotFound. Return false otherwise
func IsErrNotFound(err error) bool {
	return err == ErrHeaderNotFound
}

// IsNotErrNotFound return true if error is not nil and not equal to ErrHeaderNotFound
func IsNotErrNotFound(err error) bool {
	return !IsErrNotFound(err)
}

// GetCollectionID returns the value of the "Collection-Id" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetCollectionID(req *http.Request) (string, error) {
	return getRequestHeader(req, collectionIDHeader)
}

// GetUserAuthToken returns the value of the "X-Florence-Token" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetUserAuthToken(req *http.Request) (string, error) {
	return getRequestHeader(req, userAuthTokenHeader)
}

// GetServiceAuthToken returns the value of the "Authorization" request header if it exists, returns ErrHeaderNotFound if
// the header is not found. If the header exists the "Bearer " prefixed is removed from returned value.
func GetServiceAuthToken(req *http.Request) (string, error) {
	token, err := getRequestHeader(req, serviceAuthTokenHeader)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(token, bearerPrefix) {
		token = strings.TrimPrefix(token, bearerPrefix)
	}
	return token, nil
}

// GetDownloadServiceToken returns the value of the "X-Download-Service-Token" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetDownloadServiceToken(req *http.Request) (string, error) {
	return getRequestHeader(req, downloadServiceTokenHeader)
}

// GetUserIdentity returns the value of the "User-Identity" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetUserIdentity(req *http.Request) (string, error) {
	return getRequestHeader(req, userIdentityHeader)
}

// GetRequestID returns the value of the "X-Request-Id" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetRequestID(req *http.Request) (string, error) {
	return getRequestHeader(req, requestIDHeader)
}

// GetLocaleCode returns the value of the "LocaleCode" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetLocaleCode(req *http.Request) (string, error) {
	return getRequestHeader(req, localeCodeHeader)
}

// GetIfMatch returns the value of the "If-Match" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetIfMatch(req *http.Request) (string, error) {
	return getRequestHeader(req, ifMatchHeader)
}

// GetETag returns the value of the "ETag" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetETag(req *http.Request) (string, error) {
	return getRequestHeader(req, eTagHeader)
}

// GetResponseETag returns the value of "ETag" response header if it exists, returns
// ErrResponseNil if the header is not found.
func GetResponseETag(resp *http.Response) (string, error) {
	return getResponseHeader(resp, eTagHeader)
}

// Get Accept returns the value of the "Accept" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetAccept(req *http.Request) (string, error) {
	return getRequestHeader(req, acceptHeader)
}

func getRequestHeader(req *http.Request, headerName string) (string, error) {
	if req == nil {
		return "", ErrRequestNil
	}

	headerValue := req.Header.Get(headerName)
	if len(headerValue) == 0 {
		return "", ErrHeaderNotFound
	}

	return headerValue, nil
}

func getResponseHeader(resp *http.Response, headerName string) (string, error) {
	if resp == nil {
		return "", ErrResponseNil
	}

	headerValue := resp.Header.Get(headerName)
	if len(headerValue) == 0 {
		return "", ErrHeaderNotFound
	}

	return headerValue, nil
}

// SetCollectionID set the collection ID header on the provided request. If the collection ID header is already present
// in the request it will be overwritten by the new value. Empty values are allowed for this header
func SetCollectionID(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, collectionIDHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetServiceAuthToken set the service authentication token header on the provided request. If the authentication token is
// already present it will be overwritten by the new value. Empty values are allowed for this header.
// Replaces deprecated SetUserAuthToken function.
func SetServiceAuthToken(req *http.Request, headerValue string) error {
	if req == nil {
		return ErrRequestNil
	}

	if len(headerValue) == 0 {
		return nil
	}

	if !strings.HasPrefix(headerValue, bearerPrefix) {
		headerValue = bearerPrefix + headerValue
	}

	return setRequestHeader(req, serviceAuthTokenHeader, headerValue)
}

// SetAuthToken set the access token header on the provided request. If the access token is
// already present it will be overwritten by the new value. Empty values are allowed for this header.
func SetAuthToken(req *http.Request, headerValue string) error {
	// TODO remove the userAuthTokenHeader once the X-Florence-Token has been removed
	err := setRequestHeader(req, userAuthTokenHeader, headerValue)
	if err == ErrValueEmpty {
		return nil
	}
	if err != nil {
		return err
	}

	// Add bearer prefix if not present
	if !strings.HasPrefix(headerValue, bearerPrefix) {
		headerValue = bearerPrefix + headerValue
	}
	return setRequestHeader(req, authTokenHeader, headerValue)
}

// SetIDTokenHeader set the ID token  header on the provided request. If the authentication
// token is already present it will be overwritten by the new value. Empty values are allowed for this header.
func SetIDTokenHeader(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, idTokenHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetRefreshTokenHeader set the refresh token header on the provided request. If the authentication
// token is already present it will be overwritten by the new value. Empty values are allowed for this header.
func SetRefreshTokenHeader(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, refreshTokenHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetDownloadServiceToken set the download service auth token header on the provided request. If the authentication
// token is already present it will be overwritten by the new value. Empty values are allowed for this header.
func SetDownloadServiceToken(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, downloadServiceTokenHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetUserIdentity set the user identity header on the provided request. If a user identity token is already present it
// will be overwritten by the new value. Empty values are allowed for this header.
func SetUserIdentity(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, userIdentityHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetRequestID set the unique request ID header on the provided request. If a request ID header is already present it
// will be overwritten by the new value. Empty values are allowed for this header.
func SetRequestID(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, requestIDHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetLocaleCode set the locale code header on the provided request. If this header is already present it
// will be overwritten by the new value. Empty values are allowed for this header.
func SetLocaleCode(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, localeCodeHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

func SetAcceptedLang(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, acceptedLangHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetIfMatch set the If-Match header on the provided request. If this header is already present it
// will be overwritten by the new value. Empty values are allowed for this header.
func SetIfMatch(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, ifMatchHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetETag set the ETag header on the provided request. If this header is already present it
// will be overwritten by the new value. Empty values are allowed for this header.
func SetETag(req *http.Request, headerValue string) error {
	err := setRequestHeader(req, eTagHeader, headerValue)
	if err != nil && err != ErrValueEmpty {
		return err
	}
	return nil
}

// SetETag set the ETag header on the provided request. If this header is already present it
// will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetAccept(req *http.Request, headerValue string) error {
	return setRequestHeader(req, acceptHeader, headerValue)
}

func setRequestHeader(req *http.Request, headerName string, headerValue string) error {
	if req == nil {
		return ErrRequestNil
	}

	if len(headerValue) == 0 {
		return ErrValueEmpty
	}

	req.Header.Set(headerName, headerValue)
	return nil
}
