package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

const service = "filter-api"

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Config contains any configuration required to send requests to the filter api
type Config struct {
	InternalToken string
	FlorenceToken string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from filter api if an error is returned
func (e ErrInvalidFilterAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: filterAPIURL,
	}
}

// CloseResponseBody closes the response body and logs an error if unsuccessful
func CloseResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}

// Checker calls filter api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	hcClient := healthcheck.Client{
		Client: c.cli,
		Name:   service,
		URL:    c.url,
	}

	// healthcheck client should not retry when calling a healthcheck endpoint,
	// append to current paths as to not change the client setup by service
	paths := hcClient.Client.GetPathsWithNoRetries()
	paths = append(paths, "/health", "/healthcheck")
	hcClient.Client.SetPathsWithNoRetries(paths)

	return hcClient.Checker(ctx)
}

// Healthcheck calls the health endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()
	endpoint := "/health"

	clientlog.Do(ctx, "checking health", service, endpoint)

	resp, err := c.cli.Get(ctx, c.url+endpoint)
	// Apps may still have /healthcheck endpoint instead of a /health one.
	if resp.StatusCode == http.StatusNotFound {
		endpoint = "/healthcheck"
		resp, err = c.cli.Get(ctx, c.url+endpoint)
	}
	if err != nil {
		return service, err
	}
	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, endpoint}
	}

	return service, nil
}

// GetOutput returns a filter output job for a given filter output id
func (c *Client) GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (Model, error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s", c.url, filterOutputID)
	var m Model
	clientlog.Do(ctx, "retrieving filter output", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, uri)
	if err != nil {
		return m, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return m, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(b, &m)
	return m, err
}

// GetDimension returns information on a requested dimension name for a given filterID
func (c *Client) GetDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (Dimension, error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)
	var dim Dimension
	clientlog.Do(ctx, "retrieving dimension information", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return dim, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return dim, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dim, err
	}

	if err = json.Unmarshal(b, &dim); err != nil {
		return dim, err
	}

	return dim, err
}

// GetDimensions will return the dimensions associated with the provided filter id
func (c *Client) GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string) ([]Dimension, error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.url, filterID)
	var dims []Dimension
	clientlog.Do(ctx, "retrieving all dimensions for given filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return dims, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return dims, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dims, err
	}

	err = json.Unmarshal(b, &dims)
	return dims, err
}

// GetDimensionOptions retrieves a list of the dimension options
func (c *Client) GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) ([]DimensionOption, error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.url, filterID, name)
	var opts []DimensionOption
	clientlog.Do(ctx, "retrieving selected dimension options for filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return opts, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return opts, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return opts, err
	}

	err = json.Unmarshal(b, &opts)
	return opts, err
}

// CreateBlueprint creates a filter blueprint and returns the associated filterID
func (c *Client) CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (string, error) {
	ver, err := strconv.Atoi(version)
	if err != nil {
		return "", err
	}

	cb := CreateBlueprint{Dataset: Dataset{DatasetID: datasetID, Edition: edition, Version: ver}}

	var dimensions []ModelDimension
	for _, name := range names {
		dimensions = append(dimensions, ModelDimension{Name: name})
	}

	cb.Dimensions = dimensions

	b, err := json.Marshal(cb)
	if err != nil {
		return "", err
	}

	uri := c.url + "/filters"
	clientlog.Do(ctx, "attempting to create filter blueprint", service, uri, log.Data{
		"method":    "POST",
		"datasetID": datasetID,
		"edition":   edition,
		"version":   version,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	headers.SetDownloadServiceToken(req, downloadServiceToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return "", ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(b, &cb); err != nil {
		return "", err
	}

	return cb.FilterID, nil
}

// UpdateBlueprint will update a blueprint with a given filter model
func (c *Client) UpdateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID string, m Model, doSubmit bool) (Model, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return m, err
	}

	uri := fmt.Sprintf("%s/filters/%s", c.url, m.FilterID)

	if doSubmit {
		uri = uri + "?submitted=true"
	}

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return m, err
	}

	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	headers.SetDownloadServiceToken(req, downloadServiceToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return m, err
	}
	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return m, ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}

	return m, nil
}

// AddDimensionValue adds a particular value to a filter job for a given filterID
// and name
func (c *Client) AddDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)

	clientlog.Do(ctx, "adding dimension option to filter job", service, uri, log.Data{
		"method": "POST",
		"value":  value,
	})

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}
	return nil
}

// RemoveDimensionValue removes a particular value to a filter job for a given filterID
// and name
func (c *Client) RemoveDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.url, filterID, name, value)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}

	clientlog.Do(ctx, "removing dimension option from filter job", service, uri, log.Data{
		"method": "DELETE",
		"value":  value,
	})

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusNoContent {
		return &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
	}
	return nil
}

// RemoveDimension removes a given dimension from a filter job
func (c *Client) RemoveDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do(ctx, "removing dimension from filter job", service, uri, log.Data{
		"method":    "DELETE",
		"dimension": "name",
	})

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusNoContent {
		err = &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
		return err
	}

	return err
}

// AddDimension adds a new dimension to a filter job
func (c *Client) AddDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, name string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, id, name)
	clientlog.Do(ctx, "adding dimension to filter job", service, uri, log.Data{
		"method":    "POST",
		"dimension": name,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(`{}`))
	if err != nil {
		return err
	}
	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}

// GetJobState will return the current state of the filter job
func (c *Client) GetJobState(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) (Model, error) {
	uri := fmt.Sprintf("%s/filters/%s", c.url, filterID)
	var m Model
	clientlog.Do(ctx, "retrieving filter job state", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return m, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return m, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(b, &m)
	return m, err
}

// AddDimensionValues adds many options to a filter job dimension
func (c *Client) AddDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, options []string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)

	clientlog.Do(ctx, "adding multiple dimension values to filter job", service, uri, log.Data{
		"method":  "POST",
		"options": options,
	})

	body := struct {
		Options []string `json:"options"`
	}{
		Options: options,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	return nil
}

// GetPreview attempts to retrieve a preview for a given filterOutputID
func (c *Client) GetPreview(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (Preview, error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s/preview", c.url, filterOutputID)
	var p Preview
	clientlog.Do(ctx, "retrieving preview for filter output job", service, uri, log.Data{
		"method":   "GET",
		"filterID": filterOutputID,
	})

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, uri)
	if err != nil {
		return p, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return p, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p, err
	}

	err = json.Unmarshal(b, &p)
	return p, err
}

// doGetWithAuthHeaders executes clienter.Do setting the user and service authentication token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	return c.cli.Do(ctx, req)
}

// doGetWithAuthHeadersAndWithDownloadToken executes clienter.Do setting the user and service authentication and dwonload token token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeadersAndWithDownloadToken(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	headers.SetDownloadServiceToken(req, downloadServiceAuthToken)
	return c.cli.Do(ctx, req)
}
