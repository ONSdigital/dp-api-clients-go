// Package dimension provides an HTTP client for the Cantabular Dimension API
package dimension

import (
	"context"
	"encoding/json"
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
)

const service = "cantabular-dimension-api"

// Client is a Cantabular Dimension API client
type Client struct {
	hcCli   *health.Client
	baseURL *url.URL
}

// NewClient creates a new instance of Client with a given Dimensions API URL
func NewClient(dimensionsAPIURL string) (*Client, error) {
	client := health.NewClient(service, dimensionsAPIURL)
	return NewWithHealthClient(client)
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client
func NewWithHealthClient(hcCli *health.Client) (*Client, error) {
	client := health.NewClientWithClienter(service, hcCli.URL, hcCli.Client)
	baseURL, err := url.Parse(client.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %w", err)
	}

	return &Client{hcCli: client, baseURL: baseURL}, nil
}

// Checker calls recipe api health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// GetAreaTypes retrieves the Cantabular area-types associated with a dataset
func (c *Client) GetAreaTypes(ctx context.Context, userAuthToken, serviceAuthToken, datasetID string) (GetAreaTypesResponse, error) {
	logData := log.Data{
		"method":     http.MethodGet,
		"dataset_id": datasetID,
	}

	query := url.Values{"dataset": []string{datasetID}}.Encode()
	reqURL := c.baseURL.ResolveReference(&url.URL{Path: "/area-types", RawQuery: query}).String()

	clientlog.Do(ctx, "getting area types", service, reqURL, logData)

	req, err := newRequest(ctx, http.MethodGet, reqURL, nil, userAuthToken, serviceAuthToken)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			fmt.Errorf("failed to create request: %w", err),
			http.StatusBadRequest,
			logData,
		)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			fmt.Errorf("failed to get response from Dimensions API: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		var errorResp ErrorResp
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return GetAreaTypesResponse{}, dperrors.New(
				fmt.Errorf("error response from Dimensions API (%d): %w", resp.StatusCode, errorResp),
				http.StatusInternalServerError,
				logData,
			)
		}
	}

	if resp.StatusCode != http.StatusOK {
		// Best effort — an empty body is fine for the error message
		body, _ := io.ReadAll(resp.Body)
		return GetAreaTypesResponse{}, dperrors.New(
			fmt.Errorf("error response from Dimensions API (%d): %s", resp.StatusCode, body),
			http.StatusInternalServerError,
			logData,
		)
	}

	var areaTypes GetAreaTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&areaTypes); err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			fmt.Errorf("unable to deserialize area types response: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return areaTypes, nil
}

// newRequest creates a new http.Request with auth headers
func newRequest(ctx context.Context, method string, url string, body io.Reader, userAuthToken, serviceAuthToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set auth token header: %w", err)
	}

	if err := headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set service token header: %w", err)
	}

	return req, nil
}
