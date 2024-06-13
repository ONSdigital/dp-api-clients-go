package tablerenderer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "table-renderer"

// Client represents a dp-table-renderer client
type Client struct {
	hcCli *healthcheck.Client
}

// ErrInvalidTableRendererResponse is returned when the table-renderer service does not respond with a status 200
type ErrInvalidTableRendererResponse struct {
	responseCode int
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidTableRendererResponse) Error() string {
	return fmt.Sprintf("invalid response from table-renderer service - status %d", e.responseCode)
}

// Code returns the status code received from table-renderer if an error is returned
func (e ErrInvalidTableRendererResponse) Code() int {
	return e.responseCode
}

// New creates a new instance of Client with a given table-renderer url
func New(tableRendererURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, tableRendererURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// Checker calls table-renderer health endpoint and returns a check object to the caller.
func (r *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return r.hcCli.Checker(ctx, check)
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this cient
func (c *Client) HealthClient() *healthcheck.Client {
	return c.hcCli
}

// Render returns the given table json rendered with the given format
func (c *Client) Render(ctx context.Context, format string, json []byte) ([]byte, error) {
	if json == nil {
		json = []byte(`{}`)
	}
	return c.post(ctx, fmt.Sprintf("/render/%s", format), json)
}

func (r *Client) post(ctx context.Context, path string, b []byte) ([]byte, error) {
	uri := r.hcCli.URL + path

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidTableRendererResponse{resp.StatusCode}
	}

	return ioutil.ReadAll(resp.Body)
}
