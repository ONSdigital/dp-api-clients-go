package releasecalendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const serviceName = "release-calendar-api"

// Client is a release calendar api client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli *health.Client
}

// NewAPIClient creates a new instance of ReleaseCalendarAPI Client with a given release calendar api url
func NewAPIClient(releaseCalendarApiUrl string) *Client {
	return &Client{
		health.NewClient(serviceName, releaseCalendarApiUrl),
	}
}

// NewWithHealthClient creates a new instance of ReleaseCalendarAPI Client,
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

// HealthClient returns the underlying Healthcheck Client for this release calendar API client
func (c *Client) HealthClient() *health.Client {
	return c.hcCli
}

// Checker calls the release calendar API health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// GetLegacyRelease returns a legacy release
func (c *Client) GetLegacyRelease(ctx context.Context, userAccessToken, collectionID, lang, uri string) (*Release, error) {
	url := fmt.Sprintf("%s/releases/legacy?url=%s&lang=%s", c.hcCli.URL, uri, lang)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to create request to Release Calendar API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, err
	}
	if err = headers.SetAuthToken(req, userAccessToken); err != nil {
		return nil, err
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to get response from Release Calendar API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, c.errorResponse(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to read response body from Release Calendar API: %s", err),
			resp.StatusCode,
			nil,
		)
	}

	var release Release
	if err = json.Unmarshal(b, &release); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{"response_body": string(b)},
		)
	}

	return &release, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// errorResponse handles dealing with an error response from Release Calendar API
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
