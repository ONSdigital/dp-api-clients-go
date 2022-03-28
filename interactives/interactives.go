package interactives

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

const (
	service = "interactives-api"
)

// Client is a interactives api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// QueryParams represents the possible query parameters that a caller can provide
type QueryParams struct {
	Offset int
	Limit  int
}

func (q *QueryParams) Validate() error {
	if q.Offset < 0 || q.Limit < 0 {
		return errors.New("negative offsets or limits are not allowed")
	}
	return nil
}

// NewAPIClient creates a new instance of Client with a given interactive api url and the relevant tokens
func NewAPIClient(interactivesAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, interactivesAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// NewAPIClientWithMaxRetries creates a new instance of Client with a given interactive api url and the relevant tokens,
// setting a number of max retires for the HTTP client
func NewAPIClientWithMaxRetries(interactivesAPIURL string, maxRetries int) *Client {
	hcClient := healthcheck.NewClient(service, interactivesAPIURL)
	if maxRetries > 0 {
		hcClient.Client.SetMaxRetries(maxRetries)
	}

	return &Client{
		hcClient,
	}
}

// Checker calls interactives api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// ErrInvalidInteractivesAPIResponse is returned when the interactives api does not respond
// with a valid status
type ErrInvalidInteractivesAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

func (e ErrInvalidInteractivesAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from interactives api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// NewInteractivesAPIResponse creates an error response, optionally adding body to e when status is 404
func NewInteractivesAPIResponse(resp *http.Response, uri string) (e *ErrInvalidInteractivesAPIResponse) {
	e = &ErrInvalidInteractivesAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read InteractivesAPI body"
			return
		}
		defer closeResponseBody(nil, resp)

		e.body = string(b)
	}
	return
}

// GetInteractive returns an interactive (given the id)
func (c *Client) GetInteractive(ctx context.Context, userAuthToken, serviceAuthToken string, interactivesID string) (m Interactive, err error) {
	uri := fmt.Sprintf("%s/interactives/%s", c.hcCli.URL, interactivesID)

	clientlog.Do(ctx, fmt.Sprintf("retrieving interactive (%s)", interactivesID), service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewInteractivesAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	return
}

// ListInteractives returns the list of interactives
func (c *Client) ListInteractives(ctx context.Context, userAuthToken, serviceAuthToken string, q *QueryParams) (m List, err error) {
	uri := fmt.Sprintf("%s/interactives", c.hcCli.URL)
	if q != nil {
		if err := q.Validate(); err != nil {
			return List{}, err
		}
		uri = fmt.Sprintf("%s?offset=%d&limit=%d", uri, q.Offset, q.Limit)
	}

	clientlog.Do(ctx, "retrieving interactives", service, uri)

	resp, err := c.doListWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewInteractivesAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	return
}

// PutInteractive update the dataset
func (c *Client) PutInteractive(ctx context.Context, userAuthToken, serviceAuthToken, interactiveID string, update InteractiveUpdate) error {
	uri := fmt.Sprintf("%s/interactives/%s", c.hcCli.URL, interactiveID)

	clientlog.Do(ctx, "updating interactive", service, uri)

	payload, err := json.Marshal(update)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall interactive")
	}

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri, payload)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewInteractivesAPIResponse(resp, uri)
	}
	return nil
}

func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

func (c *Client) doListWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string, values url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if values != nil {
		req.URL.RawQuery = values.Encode()
	}

	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

func (c *Client) doPutWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string, update []byte) (*http.Response, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	if err := writer.WriteField(UpdateFormFieldKey, string(update)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, uri, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
