package search

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
	"io/ioutil"
	"net/http"
	"net/url"
)

const service = "search-query"

// ErrInvalidSearchResponse is returned when the dp-search-query does not respond
// with a valid status
type ErrInvalidSearchResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidSearchResponse) Error() string {
	return fmt.Sprintf("invalid response from dp-search-query - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from dp-search-query if an error is returned
func (e ErrInvalidSearchResponse) Code() int {
	return e.actualCode
}

// compile time check that ErrInvalidSearchResponse satisfies the error interface
var _ error = ErrInvalidSearchResponse{}

// Client is a dp-search-query client which can be used to make requests to the server
type Client struct {
	cli dphttp.Clienter
	url string
}

// NewClient creates a new instance of Client with a given search-query api url
func NewClient(searchAPIURL string) *Client {
	hcClient := healthcheck.NewClient(service, searchAPIURL)

	return &Client{
		cli: hcClient.Client,
		url: searchAPIURL,
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
	hcClient := healthcheck.Client{
		Client: c.cli,
		URL:    c.url,
		Name:   service,
	}

	return hcClient.Checker(ctx, check)
}

// doGet executes clienter.Do GET for the provided uri
// The url.Values will be added as query parameters in the URL.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	return c.cli.Do(ctx, req)
}

// NewSearchErrorResponse creates an error response
func NewSearchErrorResponse(resp *http.Response, uri string) (e *ErrInvalidSearchResponse) {
	return &ErrInvalidSearchResponse{
		expectedCode: http.StatusOK,
		actualCode: resp.StatusCode,
		uri:        uri,
	}
}

// GetSearch returns the search results
func (c *Client) GetSearch(ctx context.Context, query url.Values) (r Response, err error) {
	uri := fmt.Sprintf("%s/search", c.url)
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