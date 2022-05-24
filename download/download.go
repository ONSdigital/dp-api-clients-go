package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "download-service"

type Response struct {
	Content io.ReadCloser `json:"content"`
}

// Client is an download service client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli            *healthcheck.Client
	serviceAuthToken string
}

// NewAPIClient creates a new instance of DownloadServiceAPI Client with a given download service url
func NewAPIClient(downloadServiceAPIURL, serviceAuthToken string) *Client {
	return &Client{
		healthcheck.NewClient(service, downloadServiceAPIURL),
		serviceAuthToken,
	}
}

// NewWithHealthClient creates a new instance of DownloadServiceAPI Client,
// reusing the URL and Clienter from the provided healthcheck client.
func NewWithHealthClient(hcCli *healthcheck.Client, serviceAuthToken string) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
		serviceAuthToken,
	}
}

// Checker calls download service health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// Download returns the requested path
func (c *Client) Download(ctx context.Context, path string) (*Response, error) {
	uri := fmt.Sprintf("%s/downloads-new/%s", c.hcCli.URL, path)

	clientlog.Do(ctx, "retrieving resource", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, uri)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to create request to DownloadService API: %w", err),
			http.StatusInternalServerError,
			nil,
		)
	}

	//resp.StatusCode == http.StatusMovedPermanently - redirects followed by default with httpclient: https://github.com/golang/go/issues/42832

	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, dperrors.New(
				fmt.Errorf("failed to read error response body: %s", err),
				resp.StatusCode,
				nil,
			)
		}
		closeResponseBody(ctx, resp)
		return nil, dperrors.New(
			errors.New(string(b)), resp.StatusCode, nil,
		)
	}

	return &Response{Content: resp.Body}, nil
}

func (c *Client) doGetWithAuthHeaders(ctx context.Context, uri string) (*http.Response, error) {
	clientlog.Do(ctx, "retrieving resource", service, uri)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	dprequest.AddServiceTokenHeader(req, c.serviceAuthToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to create request to DownloadService API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}
	return resp, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
