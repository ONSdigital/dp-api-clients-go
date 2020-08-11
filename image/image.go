package image

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
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

// Client is an image api client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli *healthcheck.Client
}

// closeResponseBody closes the response body and logs an error containing the context if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}

// NewAPIClient creates a new instance of ImageAPI Client with a given image api url
func NewAPIClient(imageAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, imageAPIURL),
	}
}

// NewWithHealthClient creates a new instance of ImageAPI Client,
// reusing the URL and Clienter from the provided healthcheck client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this image API client
func (c *Client) HealthClient() *healthcheck.Client {
	return c.hcCli
}

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// GetImages returns the list of images
func (c *Client) GetImages(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string) (m Images, err error) {
	uri := fmt.Sprintf("%s/images", c.hcCli.URL)

	clientlog.Do(ctx, "retrieving images", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
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
func (c *Client) PostImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, data NewImage) (m Image, err error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/images", c.hcCli.URL)

	clientlog.Do(ctx, "posting new image", service, uri)

	resp, err := c.doPostWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
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
func (c *Client) GetImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string) (m Image, err error) {
	uri := fmt.Sprintf("%s/images/%s", c.hcCli.URL, imageID)

	clientlog.Do(ctx, "retrieving images", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
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
func (c *Client) PutImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string, data Image) (m Image, err error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/images/%s", c.hcCli.URL, imageID)

	clientlog.Do(ctx, "updating instance import_tasks", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
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

// GetDownloadVariants returns the list of download variants for an image
func (c *Client) GetDownloadVariants(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string) (m ImageDownloads, err error) {
	uri := fmt.Sprintf("%s/images/%s/downloads", c.hcCli.URL, imageID)

	clientlog.Do(ctx, "retrieving download variants", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
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

// GetDownloadVariant returns a requested download variant for an image
func (c *Client) GetDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string) (m ImageDownload, err error) {
	uri := fmt.Sprintf("%s/images/%s/downloads/%s", c.hcCli.URL, imageID, variant)

	clientlog.Do(ctx, "retrieving download variant", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
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

// PutDownloadVariant updates the specified download variant for the specified image
func (c *Client) PutDownloadVariant(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID, variant string, data ImageDownload) (m ImageDownload, err error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("%s/images/%s/downloads/%s", c.hcCli.URL, imageID, variant)

	clientlog.Do(ctx, "updating image download variant", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
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

// PublishImage triggers an image publishing
func (c *Client) PublishImage(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, imageID string) (err error) {

	uri := fmt.Sprintf("%s/images/%s/publish", c.hcCli.URL, imageID)

	clientlog.Do(ctx, "publishing image", service, uri)

	resp, err := c.doPostWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, []byte{})
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewImageAPIResponse(resp, uri)
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

func addCollectionIDHeader(r *http.Request, collectionID string) {
	if len(collectionID) > 0 {
		r.Header.Add(common.CollectionIDHeaderKey, collectionID)
	}
}

// doGetWithAuthHeaders executes clienter.Do GET for the provided uri, setting the required headers according to the provided useAuthToken, serviceAuthToken and collectionID.
// If url.Values are provided, they will be added as query parameters in the URL.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, values url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if values != nil {
		req.URL.RawQuery = values.Encode()
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// doPostWithAuthHeaders executes clienter.Do POST for the provided uri, setting the required headers according to the provided useAuthToken, serviceAuthToken and collectionID.
// The provided payload byte array will be sent as request body.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doPostWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// doPutWithAuthHeaders executes clienter.Do PUT for the provided uri, setting the required headers according to the provided useAuthToken, serviceAuthToken and collectionID.
// The provided payload byte array will be sent as request body.
// Returns the http.Response and any error and it is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doPutWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	common.AddFlorenceHeader(req, userAuthToken)
	common.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}
