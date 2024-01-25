package berlin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin/models"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

const (
	service = "dp-nlp-berlin-api"
)

type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given berlin api url
func New(berlinAPIURL string) *Client {
	return &Client{
		hcCli: healthcheck.NewClient(service, berlinAPIURL),
	}
}

// NewWithHealthClient creates a new instance of berlin API Client,
// reusing the URL and Clienter from the provided healthcheck client
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		hcCli: healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (cli *Client) URL() string {
	return cli.hcCli.URL
}

// Health returns the underlying Healthcheck Client for this berlin API client
func (cli *Client) Health() *healthcheck.Client {
	return cli.hcCli
}

// Checker calls berlin api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return cli.hcCli.Checker(ctx, check)
}

// GetBerlin gets a list of berlin results based on the berlin request
func (cli *Client) GetBerlin(ctx context.Context, options Options) (*models.Berlin, errors.Error) {
	path := fmt.Sprintf("%s/berlin/search", cli.URL())
	if options.Query != nil {
		path = path + "?" + options.Query.Encode()
	}

	respInfo, apiErr := cli.callBerlinAPI(ctx, path, http.MethodGet, options.Headers, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var berlinResp models.Berlin

	if err := json.Unmarshal(respInfo.Body, &berlinResp); err != nil {
		return nil, errors.StatusError{
			Err: fmt.Errorf("failed to unmarshal berlin response - error is: %v", err),
		}
	}

	return &berlinResp, nil
}

type ResponseInfo struct {
	Body    []byte
	Headers http.Header
	Status  int
}

// callBerlinAPI calls the Berlin API endpoint given by path for the provided REST method, request headers, and body payload.
// It returns the response body and any error that occurred.
func (cli *Client) callBerlinAPI(ctx context.Context, path, method string, headers http.Header, payload []byte) (*ResponseInfo, errors.Error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, errors.StatusError{
			Err: fmt.Errorf("failed to parse path: \"%v\" error is: %v", path, err),
		}
	}

	path = URL.String()

	var body io.Reader

	if payload != nil {
		body = bytes.NewReader(payload)
	} else {
		body = http.NoBody
	}

	req, err := http.NewRequest(method, path, body)

	// check req, above, didn't error
	if err != nil {
		return nil, errors.StatusError{
			Err: fmt.Errorf("failed to create request for call to berlin api, error is: %v", err),
		}
	}

	// set any headers against request
	setHeaders(req, headers)

	if payload != nil {
		req.Header.Add("Content-type", "application/json")
	}

	resp, err := cli.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, errors.StatusError{
			Err:  fmt.Errorf("failed to call berlin api, error is: %v", err),
			Code: http.StatusInternalServerError,
		}
	}
	defer func() {
		err = closeResponseBody(resp)
	}()

	respInfo := &ResponseInfo{
		Headers: resp.Header.Clone(),
		Status:  resp.StatusCode,
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 400 {
		return respInfo, errors.StatusError{
			Err:  fmt.Errorf("failed as unexpected code from berlin api: %v", resp.StatusCode),
			Code: resp.StatusCode,
		}
	}

	if resp.Body == nil {
		return respInfo, nil
	}

	respInfo.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return respInfo, errors.StatusError{
			Err:  fmt.Errorf("failed to read response body from call to berlin api, error is: %v", err),
			Code: resp.StatusCode,
		}
	}
	return respInfo, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(resp *http.Response) errors.Error {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			return errors.StatusError{
				Err:  fmt.Errorf("error closing http response body from call to berlin api, error is: %v", err),
				Code: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
