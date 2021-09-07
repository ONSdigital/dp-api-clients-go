package recipe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "recipe-api"

// Client is a recpie api client which can be used to make requests to the server
type Client struct {
	hcCli *health.Client
}

// NewClient creates a new instance of Client with a given recipe api url
func NewClient(recipeAPIURL string) *Client {
	return &Client{
		health.NewClient(service, recipeAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *health.Client) *Client {
	return &Client{
		health.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls recipe api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// doGetWithAuthHeaders executes clienter.Do setting the provided user and service auth tokens as headers.
// Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	err = headers.SetAuthToken(req, userAuthToken)
	if err != nil {
		return nil, err
	}
	err = headers.SetServiceAuthToken(req, serviceAuthToken)
	if err != nil {
		return nil, err
	}
	return c.hcCli.Client.Do(ctx, req)
}

// errorResponse handles dealing with an error response from Recipe API
func (c *Client) errorResponse(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		dperrors.New(
			fmt.Errorf("failed to read error response body: %s", err),
			res.StatusCode,
			nil,
		)
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var resp ErrorResponse

	if err := json.Unmarshal(b, &resp); err != nil {
		dperrors.New(
			fmt.Errorf("failed to unmarshal error response body: %s", err),
			res.StatusCode,
			log.Data{
				"response_body": string(b),
			},
		)
	}

	return dperrors.New(
		errors.New(resp.Message),
		res.StatusCode,
		nil,
	)
}
