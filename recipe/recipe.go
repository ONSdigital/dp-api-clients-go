package recipe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
)

const service = "recipe-api"

// Client is a recpie api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given recipe api url
func New(recipeAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, recipeAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls recipe api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// CloseResponseBody closes the response body
func CloseResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		resp.Body.Close()
	}
}

// GetRecipe from an ID
func (c *Client) GetRecipe(ctx context.Context, userAuthToken, serviceAuthToken, recipeID string) (*Recipe, error) {
	uri := fmt.Sprintf("%s/recipes/%s", c.hcCli.URL, recipeID)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to call recipe api: %w", err)
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &Error{
			err:        errors.New("wrong status code, expected 200 OK"),
			statusCode: resp.StatusCode,
			logData:    log.Data{},
		}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &Error{
			err:        fmt.Errorf("failed to read response from recipe-api: %s", err),
			statusCode: resp.StatusCode,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	var recipe Recipe
	if err = json.Unmarshal(b, &recipe); err != nil {
		return nil, &Error{
			err:        fmt.Errorf("failed to unmarshal response from recipe-api: %s", err),
			statusCode: resp.StatusCode,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	return &recipe, nil
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
