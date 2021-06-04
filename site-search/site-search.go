package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
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

// closeResponseBody closes the response body and logs an error containing the context if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}

// Checker calls search api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// doGet executes clienter.Do GET for the provided uri
// The url.Values will be added as query parameters in the URL.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	return c.hcCli.Client.Do(ctx, req)
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
func (c *Client) GetSearch(ctx context.Context, query url.Values) (r Response, err error) {
	uri := fmt.Sprintf("%s/search", c.hcCli.URL)
	if query != nil {
		uri = uri + "?" + query.Encode()
	}

	clientlog.Do(ctx, "retrieving search response", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, uri)
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
func (c *Client) GetDepartments(ctx context.Context, query url.Values) (d Department, err error) {
	uri := fmt.Sprintf("%s/departments/search", c.hcCli.URL)
	if query != nil {
		uri = uri + "?" + query.Encode()
	}

	clientlog.Do(ctx, "retrieving departments search response", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, uri)
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
