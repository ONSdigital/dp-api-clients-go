package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "search-api"

// ErrInvalidSearchResponse is returned when the dp-search-api does not respond
// with a valid status
type ErrInvalidSearchResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidSearchResponse) Error() string {
	return fmt.Sprintf("invalid response from dp-search-api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from dp-search-api if an error is returned
func (e ErrInvalidSearchResponse) Code() int {
	return e.actualCode
}

// compile time check that ErrInvalidSearchResponse satisfies the error interface
var _ error = ErrInvalidSearchResponse{}

// Client is a dp-search-api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// NewClient creates a new instance of Client with a given search-api url
func NewClient(searchAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, searchAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// Checker calls search api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// doGet executes clienter.Do GET for the provided uri
// The url.Values will be added as query parameters in the URL.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

func addCollectionIDHeader(r *http.Request, collectionID string) {
	if len(collectionID) > 0 {
		r.Header.Add(dprequest.CollectionIDHeaderKey, collectionID)
	}
}

// NewSearchErrorResponse creates an error response
func NewSearchErrorResponse(resp *http.Response, uri string) (e *ErrInvalidSearchResponse) {
	return &ErrInvalidSearchResponse{
		expectedCode: http.StatusOK,
		actualCode:   resp.StatusCode,
		uri:          uri,
	}
}

// GetSearch returns the search results
func (c *Client) GetSearch(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (r Response, err error) {
	uri := fmt.Sprintf("%s/search", c.hcCli.URL)
	if query != nil {
		uri = uri + "?" + query.Encode()
	}

	clientlog.Do(ctx, "retrieving search response", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewSearchErrorResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &r); err != nil {
		return
	}

	return
}

// GetDepartments returns the search results
func (c *Client) GetDepartments(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (d Department, err error) {
	uri := fmt.Sprintf("%s/departments/search", c.hcCli.URL)
	if query != nil {
		uri = uri + "?" + query.Encode()
	}

	clientlog.Do(ctx, "retrieving departments search response", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewSearchErrorResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &d); err != nil {
		return
	}

	return
}

// GetReleases returns the search results for published Releases and upcoming Release Calendar entries
func (c *Client) GetReleases(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (ReleaseResponse, error) {
	uri := fmt.Sprintf("%s/search/releases", c.hcCli.URL)
	if query != nil {
		uri = uri + "?" + query.Encode()
	}
	clientlog.Do(ctx, "retrieving releases search response", service, uri)

	var r ReleaseResponse
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return r, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewSearchErrorResponse(resp, uri)
		return r, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	if err = json.Unmarshal(b, &r); err != nil {
		return r, err
	}

	return r, nil
}
