package upload

import (
	"context"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
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

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}
