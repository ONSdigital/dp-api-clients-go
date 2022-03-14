package filterflex

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/pkg/errors"
)

const service = "cantabular-filter-flex-api"

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	health *health.Client
	cfg    Config
}

// New creates a new instance of Client with a given host api url
func New(cfg Config) *Client {
	return &Client{
		cfg:    cfg,
		health: health.NewClient(service, cfg.HostURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(cfg Config, cli *health.Client) *Client {
	return &Client{
		health: health.NewClientWithClienter(service, cli.URL, cli.Client),
		cfg:    cfg,
	}
}

// Checker calls filter api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.health.Checker(ctx, check)
}

// ForwardRequest is used for forwarding a request from another service. Initially
// implemented for fowarding requests for Cantabular based datasets from dp-filter-api.
// The provided request is expected have any required headers as the original request
// will have been made using the relevant api-client. Note that the caller is responsible
// for closing the response body as with making a regular http request.
func (c *Client) ForwardRequest(req *http.Request) (*http.Response, error) {
	parsedHostURL, err := url.Parse(c.cfg.HostURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config host url")
	}
	parsedHostURL.Path = req.URL.Path
	parsedHostURL.RawQuery = req.URL.Query().Encode()

	proxyReq, err := http.NewRequest(req.Method, parsedHostURL.String(), req.Body)
	if err != nil {
		return nil, &Error{
			err: errors.Wrap(err, "failed to create proxy request"),
			logData: log.Data{
				"target_uri":     parsedHostURL.String(),
				"request_method": req.Method,
			},
		}
	}

	proxyReq.Header = req.Header

	return c.health.Client.Do(req.Context(), proxyReq)
}
