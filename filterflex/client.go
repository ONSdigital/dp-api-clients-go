package filterflex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
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

// newRequest creates a new http.Request with auth headers
func newRequest(ctx context.Context, method string, url string, body io.Reader, userAuthToken, serviceAuthToken, ifMatch string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	if err := headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, errors.Wrap(err, "failed to set auth token header")
	}

	if err := headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, errors.Wrap(err, "failed to set service token header")
	}

	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return nil, fmt.Errorf("failed to set if match: %w", err)
	}
	return req, nil
}

func (c *Client) createDeleteOptionRequest(ctx context.Context, input GetDeleteOptionInput) (*http.Request, error) {
	parsedHostURL, err := url.Parse(c.cfg.HostURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config host url")
	}

	urlPath := fmt.Sprintf("/filters/%s/dimensions/%s/options/%s", input.FilterID, input.Dimension, input.Option)
	urlValues := url.Values{}

	parsedHostURL.Path = urlPath
	parsedHostURL.RawQuery = urlValues.Encode()
	reqURL := parsedHostURL.String()

	req, err := newRequest(ctx, http.MethodPost, reqURL, nil, input.UserAuthToken, input.ServiceAuthToken, input.IfMatch)
	if err != nil {
		return &http.Request{}, dperrors.New(
			errors.Wrap(err, "failed to create request"),
			http.StatusBadRequest,
			log.Data{},
		)
	}
	return req, nil
}

func (c *Client) DeleteOption(ctx context.Context, input GetDeleteOptionInput) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s", c.health.URL, input.FilterID)
	clientlog.Do(ctx, "retrieving filter", service, uri)

	logData := log.Data{
		"method":    http.MethodPost,
		"filter_id": input.FilterID,
		"dimension": input.Dimension,
		"option":    input.Option,
		"ifMatch":   input.IfMatch,
	}

	req, err := c.createDeleteOptionRequest(ctx, input)
	if err != nil {
		return "", dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "deleting options", service, req.URL.String(), logData)

	resp, err := c.health.Client.Do(ctx, req)
	if err != nil {
		return "", dperrors.New(
			errors.Wrap(err, "failed to get response from filter flex API"),
			http.StatusInternalServerError,
			logData,
		)
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusNoContent {
		return "", dperrors.New(
			errors.Wrap(err, "failed to get response from filter flex API"),
			resp.StatusCode,
			logData,
		)
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
