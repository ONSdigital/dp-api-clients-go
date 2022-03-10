package upload

import (
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
)

const service = "upload-api"

// Client is an upload API client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli *healthcheck.Client
}

// NewAPIClient creates a new instance of Upload Client with a given image API URL
func NewAPIClient(uploadAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, uploadAPIURL),
	}
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this image API client
func (c *Client) HealthClient() *healthcheck.Client {
	return c.hcCli
}

// NewWithHealthClient creates a new instance of ImageAPI Client,
// reusing the URL and Clienter from the provided healthcheck client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}
