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
	dprequest "github.com/ONSdigital/dp-net/request"
)

const (
	service       = "dimension-search-api"
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

// ErrInvalidDimensionSearchAPIResponse is returned when the dimension-search api does not respond
// with a valid status
type ErrInvalidDimensionSearchAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidDimensionSearchAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from dimension-search api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from dimension-search api if an error is returned
func (e ErrInvalidDimensionSearchAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidDimensionSearchAPIResponse{}

// Client is a search api client that can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given dimension-search api url
func New(dimensionSearchAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, dimensionSearchAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls dimension-search api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
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

	uri := fmt.Sprintf("%s/dimension-search/datasets/%s/editions/%s/versions/%s/dimensions/%s?",
		c.hcCli.URL,
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
			req.Header.Set(dprequest.DeprecatedAuthHeader, params[0].InternalToken)
		}
		if len(params[0].FlorenceToken) > 0 {
			req.Header.Set(dprequest.FlorenceHeaderKey, params[0].FlorenceToken)
		}
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidDimensionSearchAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	err = json.NewDecoder(resp.Body).Decode(&m)

	return
}
