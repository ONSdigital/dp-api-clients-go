package population

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "population-types-api"

// Client is a Cantabular Population Types API client
type Client struct {
	hcCli   *health.Client
	baseURL *url.URL
}

// NewClient creates a new instance of Client with a given Population Type API URL
func NewClient(apiURL string) (*Client, error) {
	client := health.NewClient(service, apiURL)
	return NewWithHealthClient(client)
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client
func NewWithHealthClient(hcCli *health.Client) (*Client, error) {
	client := health.NewClientWithClienter(service, hcCli.URL, hcCli.Client)
	baseURL, err := url.Parse(client.URL)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing URL")
	}

	// The Parse method on `url.URL` uses a trailing slash to determine
	// how relative URLs are joined.
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path = baseURL.Path + "/"
	}

	return &Client{hcCli: client, baseURL: baseURL}, nil
}

// Checker calls recipe api health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx context.Context, check *healthcheck.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

func (c *Client) createGetRequest(ctx context.Context, userAuthToken, serviceAuthToken, urlPath string, urlValues url.Values) (*http.Request, error) {
	areasURL, err := c.baseURL.Parse(urlPath)
	if err != nil {
		return &http.Request{}, dperrors.New(
			errors.Wrap(err, "failed to parse areas URL"),
			http.StatusInternalServerError,
			log.Data{},
		)
	}

	areasURL.RawQuery = urlValues.Encode()
	reqURL := areasURL.String()

	req, err := newRequest(ctx, http.MethodGet, reqURL, nil, userAuthToken, serviceAuthToken)
	if err != nil {
		return &http.Request{}, dperrors.New(
			errors.Wrap(err, "failed to create request"),
			http.StatusBadRequest,
			log.Data{},
		)
	}
	return req, nil
}

func checkGetResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read error response body: %w", err)
		}

		var errorResp ErrorResp
		if err := json.Unmarshal(b, &errorResp); err != nil {
			return dperrors.New(
				fmt.Errorf("failed to unmarshal response body: %w", err),
				resp.StatusCode,
				log.Data{
					"response_body": string(b),
				},
			)
		}

		return dperrors.New(
			fmt.Errorf("error response from Population Type API: %w", errorResp),
			resp.StatusCode,
			nil,
		)
	}

	return nil
}

// newRequest creates a new http.Request with auth headers
func newRequest(ctx context.Context, method string, url string, body io.Reader, userAuthToken, serviceAuthToken string) (*http.Request, error) {
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

	return req, nil
}
