package recipe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
)

const service = "recipe-api"

// Client is a recpie api client which can be used to make requests to the server
type Client struct {
	hcCli *health.Client
}

// New creates a new instance of Client with a given recipe api url
func New(recipeAPIURL string) *Client {
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

// closeResponseBody closes the response body
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		resp.Body.Close()
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

	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// errorResponse handles dealing with an error response from Recipe API
func (c *Client) errorResponse(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &Error{
			err:        fmt.Errorf("failed to read error response body: %s", err),
			statusCode: res.StatusCode,
		}
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var resp ErrorResponse

	if err := json.Unmarshal(b, &resp); err != nil {
		return &Error{
			err:        fmt.Errorf("failed to unmarshal error response body: %s", err),
			statusCode: res.StatusCode,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	return &Error{
		err:        errors.New(resp.Message),
		statusCode: res.StatusCode,
	}
}
