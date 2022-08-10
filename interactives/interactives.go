package interactives

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const (
	service  = "interactives-api"
	rootPath = "interactives"
)

// Client is a interactives api client which can be used to make requests to the server
type Client struct {
	hcCli   *healthcheck.Client
	version string
}

// NewAPIClient creates a new instance of Client with a given interactive api url and the relevant tokens
func NewAPIClient(interactivesAPIURL, version string) *Client {
	return &Client{
		healthcheck.NewClient(service, interactivesAPIURL), version,
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client, version string) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client), version,
	}
}

// NewAPIClientWithMaxRetries creates a new instance of Client with a given interactive api url and the relevant tokens,
// setting a number of max retires for the HTTP client
func NewAPIClientWithMaxRetries(interactivesAPIURL, version string, maxRetries int) *Client {
	hcClient := healthcheck.NewClient(service, interactivesAPIURL)
	if maxRetries > 0 {
		hcClient.Client.SetMaxRetries(maxRetries)
	}

	return &Client{
		hcClient, version,
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
	uri := fmt.Sprintf("%s/%s/%s/%s", c.hcCli.URL, c.version, rootPath, interactivesID)

	clientlog.Do(ctx, fmt.Sprintf("retrieving interactive (%s)", interactivesID), service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri, nil)
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
func (c *Client) ListInteractives(ctx context.Context, userAuthToken, serviceAuthToken string, filter *Filter) (m []Interactive, err error) {
	uri := fmt.Sprintf("%s/%s/%s", c.hcCli.URL, c.version, rootPath)
	var qVals url.Values
	if filter != nil {
		qVals = url.Values{}
		marshalled, jsonerr := json.Marshal(filter)
		if jsonerr != nil {
			return []Interactive{}, jsonerr
		}
		qVals["filter"] = []string{string(marshalled)}
	}

	clientlog.Do(ctx, "retrieving interactives", service, uri, log.Data{"query_params": qVals})

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri, qVals)
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

// PutInteractive update the interactive
func (c *Client) PutInteractive(ctx context.Context, userAuthToken, serviceAuthToken, interactiveID string, interactive Interactive) error {
	uri := fmt.Sprintf("%s/%s/%s/%s", c.hcCli.URL, c.version, rootPath, interactiveID)

	clientlog.Do(ctx, "updating interactive", service, uri)

	payload, err := json.Marshal(interactive)
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

// PatchInteractive patches the interactive
func (c *Client) PatchInteractive(ctx context.Context, userAuthToken, serviceAuthToken, interactiveID string, req PatchRequest) (i Interactive, err error) {
	uri := fmt.Sprintf("%s/%s/%s/%s", c.hcCli.URL, c.version, rootPath, interactiveID)

	clientlog.Do(ctx, "patching interactive", service, uri)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return i, errors.Wrap(err, "error while attempting to marshall interactive")
	}

	resp, err := c.doPatchWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri, buf)
	if err != nil {
		return i, errors.Wrap(err, "http client returned error while attempting to make request")
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
	if err = json.Unmarshal(b, &i); err != nil {
		return
	}

	return
}

func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string, values url.Values) (*http.Response, error) {
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

func (c *Client) doPatchWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, uri string, buf *bytes.Buffer) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, uri, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
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
