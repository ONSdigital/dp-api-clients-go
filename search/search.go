package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

const (
	service       = "search-api"
	defaultLimit  = 50
	defaultOffset = 0
)

// Config represents configuration required to conduct a search request
type Config struct {
	Limit         *int
	Offset        *int
	InternalToken string
	FlorenceToken string
}

// ErrInvalidSearchAPIResponse is returned when the search api does not respond
// with a valid status
type ErrInvalidSearchAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidSearchAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from search api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from search api if an error is returned
func (e ErrInvalidSearchAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidSearchAPIResponse{}

// Client is a search api client that can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
}

// New creates a new instance of Client with a given search api url
func New(searchAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: searchAPIURL,
	}
}

// Checker calls search api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	hcClient := healthcheck.Client{
		Client: c.cli,
		Name:   service,
		URL:    c.url,
	}

	// healthcheck client should not retry when calling a healthcheck endpoint,
	// append to current paths as to not change the client setup by service
	paths := hcClient.Client.GetPathsWithNoRetries()
	paths = append(paths, "/health", "/healthcheck")
	hcClient.Client.SetPathsWithNoRetries(paths)

	return hcClient.Checker(ctx)
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()
	endpoint := "/health"

	clientlog.Do(ctx, "checking health", service, endpoint)

	resp, err := c.cli.Get(ctx, c.url+endpoint)
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	// Apps may still have /healthcheck endpoint instead of a /health one.
	if resp.StatusCode == http.StatusNotFound {
		endpoint = "/healthcheck"
		return c.callHealthcheckEndpoint(ctx, service, endpoint)
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, c.url + endpoint}
	}

	return service, nil
}

// Dimension allows the searching of a dimension for a specific dimension option, optionally
// pass in configuration parameters as an additional field. This can include a request specific
// internal token
func (c *Client) Dimension(ctx context.Context, datasetID, edition, version, name, query string, params ...Config) (m *Model, err error) {
	offset := defaultOffset
	limit := defaultLimit

	if len(params) > 0 {
		if params[0].Offset != nil {
			offset = *params[0].Offset
		}
		if params[0].Limit != nil {
			limit = *params[0].Limit
		}
	}

	uri := fmt.Sprintf("%s/search/datasets/%s/editions/%s/versions/%s/dimensions/%s?",
		c.url,
		datasetID,
		edition,
		version,
		name,
	)

	v := url.Values{}
	v.Add("q", query)
	v.Add("limit", strconv.Itoa(limit))
	v.Add("offset", strconv.Itoa(offset))

	uri = uri + v.Encode()

	clientlog.Do(ctx, "searching for dataset dimension option", service, uri)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return
	}

	if len(params) > 0 {
		if len(params[0].InternalToken) > 0 {
			req.Header.Set(common.DeprecatedAuthHeader, params[0].InternalToken)
		}
		if len(params[0].FlorenceToken) > 0 {
			req.Header.Set(common.FlorenceHeaderKey, params[0].FlorenceToken)
		}
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	err = json.NewDecoder(resp.Body).Decode(&m)

	return
}

func (c *Client) callHealthcheckEndpoint(ctx context.Context, service, endpoint string) (string, error) {
	clientlog.Do(ctx, "checking health", service, endpoint)
	resp, err := c.cli.Get(ctx, c.url+endpoint)
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidSearchAPIResponse{http.StatusOK, resp.StatusCode, c.url + endpoint}
	}

	return service, nil
}

// CloseResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}
