package cantabular

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	dperrors "github.com/ONSdigital/dp-api-clients-go/errors"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "cantabular"

// Client is the client for interacting with the Cantabular API
type Client struct {
	ua   httpClient
	host string
}

// NewClient returns a new Client
func NewClient(ua httpClient, cfg Config) *Client {
	return &Client{
		ua:   ua,
		host: cfg.Host,
	}
}

// get makes a get request to the given url and returns the response
func (c *Client) httpGet(ctx context.Context, path string) (*http.Response, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to parse url: %s", err),
			http.StatusBadRequest,
			log.Data{
				"url": path,
			},
		)
	}

	path = URL.String()

	resp, err := c.ua.Get(ctx, path)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make request: %w", err),
			http.StatusInternalServerError,
			log.Data{
				"url": path,
			},
		)
	}

	return resp, nil
}

// Checker is a placeholder for Cantabular healthcheck, it returns StatusOK with a 'not implemented' message
// TODO implement this function when we decide how to check the health status of a cantabular service
func (c *Client) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	code := 0
	state.Update(healthcheck.StatusOK, "cantabular healthcheck not implemendted", code)
	return nil
}

// errorResponse handles dealing with an error response from Cantabular
func (c *Client) errorResponse(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read error response body: %s", err),
			res.StatusCode,
			nil,
		)
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var resp ErrorResponse

	if err := json.Unmarshal(b, &resp); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal error response body: %s", err),
			res.StatusCode,
			log.Data{
				"response_body": string(b),
			},
		)
	}

	return dperrors.New(
		errors.New(resp.Message),
		res.StatusCode,
		nil,
	)
}
