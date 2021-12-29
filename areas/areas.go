package areas

import (
	"context"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const service = "areas-api"


// Client is a areas api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}


// New creates a new instance of Client with a given areas api url
func New(areasAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, areasAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls areas api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}
