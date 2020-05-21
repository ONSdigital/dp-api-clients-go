package image

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/log"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	dphttp "github.com/ONSdigital/dp-net/http"
)

const service = "image-api"

// ErrInvalidImageAPIResponse is returned when the image api does not respond
// with a valid status
type ErrInvalidImageAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidImageAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from image api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from image api if an error is returned
func (e ErrInvalidImageAPIResponse) Code() int {
	return e.actualCode
}

// compile time check that ErrInvalidImageAPIResponse satisfies the error interface
var _ error = ErrInvalidImageAPIResponse{}

// Client is a image api client which can be used to make requests to the server
type Client struct {
	cli dphttp.Clienter
	url string
}

// closeResponseBody closes the response body and logs an error containing the context if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}

// NewAPIClient creates a new instance of Client with a given image api url and the relevant tokens
func NewAPIClient(imageAPIURL string) *Client {
	hcClient := healthcheck.NewClient(service, imageAPIURL)

	return &Client{
		cli: hcClient.Client,
		url: imageAPIURL,
	}
}

// NewAPIClientWithMaxRetries creates a new instance of Client with a given image api url and the relevant tokens,
// setting a number of max retires for the HTTP client
func NewAPIClientWithMaxRetries(imageAPIURL string, maxRetries int) *Client {
	hcClient := healthcheck.NewClient(service, imageAPIURL)
	if maxRetries > 0 {
		hcClient.Client.SetMaxRetries(maxRetries)
	}

	return &Client{
		cli: hcClient.Client,
		url: imageAPIURL,
	}
}

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	hcClient := healthcheck.Client{
		Client: c.cli,
		URL:    c.url,
		Name:   service,
	}

	return hcClient.Checker(ctx, check)
}

// GetImages returns the list of images
func (c *Client) GetImages(ctx context.Context) (m Images, err error) {
	uri := fmt.Sprintf("%s/images", c.url)

	clientlog.Do(ctx, "retrieving images", service, uri)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewImageAPIResponse(resp, uri)
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

// PostImage performs a 'POST /images' with the provided NewImage
func (c *Client) PostImage(ctx context.Context, data NewImage) (m Image, err error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/images", c.url)

	clientlog.Do(ctx, "posting new image", service, uri)

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(payload))
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewImageAPIResponse(resp, uri)
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

// GetImage returns a requested image
func (c *Client) GetImage(ctx context.Context, imageID string) (m Image, err error) {
	uri := fmt.Sprintf("%s/images/%s", c.url, imageID)

	clientlog.Do(ctx, "retrieving images", service, uri)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewImageAPIResponse(resp, uri)
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

// PutImage updates the specified image
func (c *Client) PutImage(ctx context.Context, imageID string, data Image) (m Image, err error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/images/%s", c.url, imageID)

	clientlog.Do(ctx, "updating instance import_tasks", service, uri)
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewReader(payload))
	if err != nil {
		return
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewImageAPIResponse(resp, uri)
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

// NewImageAPIResponse creates an error response, optionally adding body to e when status is 404
func NewImageAPIResponse(resp *http.Response, uri string) (e *ErrInvalidImageAPIResponse) {
	e = &ErrInvalidImageAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read ImageAPI body"
			return
		}
		defer closeResponseBody(nil, resp)

		e.body = string(b)
	}
	return
}
