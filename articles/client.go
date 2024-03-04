package articles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const serviceName = "articles-api"

// Client is an articles api client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli *health.Client
}

// NewAPIClient creates a new instance of ArticlesApi Client with a given article api url
func NewAPIClient(articlesAPIURL string) *Client {
	return &Client{
		health.NewClient(serviceName, articlesAPIURL),
	}
}

// NewWithHealthClient creates a new instance of ArticlesApi Client,
// reusing the URL and Clienter from the provided healthcheck client.
func NewWithHealthClient(hcCli *health.Client) *Client {
	return &Client{
		health.NewClientWithClienter(serviceName, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this articles API client
func (c *Client) HealthClient() *health.Client {
	return c.hcCli
}

// Checker calls articles API health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// errorResponse handles dealing with an error response from Articles API
func (c *Client) errorResponse(res *http.Response) error {
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read error response body: %s", err),
			res.StatusCode,
			nil,
		)
	}

	return dperrors.New(
		errors.New(string(b)),
		res.StatusCode,
		nil,
	)
}
