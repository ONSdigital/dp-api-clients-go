package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/batch"
	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "filter-api"

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	ExpectedCode int
	ActualCode   int
	URI          string
}

// error definitions that are not related to invalid responses
var (
	ErrBatchETagMismatch      = errors.New("ETag value changed from one batch to another")
	ErrBatchUnexpectedType    = errors.New("batch processor was called with an unexpected type of items")
	ErrInvalidPaginationQuery = errors.New("negative offsets or limits are not allowed")
)

// Config contains any configuration required to send requests to the filter api
type Config struct {
	InternalToken string
	FlorenceToken string
}

// DimensionOptionsBatchProcessor is the type corresponding to a batch processing function for filter DimensionOptions
type DimensionOptionsBatchProcessor func(opts DimensionOptions, eTag string) (abort bool, err error)

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.ExpectedCode,
		e.ActualCode,
		e.URI,
	)
}

// Code returns the status code received from filter api if an error is returned
func (e ErrInvalidFilterAPIResponse) Code() int {
	return e.ActualCode
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// QueryParams represents the possible query parameters that a caller can provide
type QueryParams struct {
	Offset int
	Limit  int
}

// Validate validates that no negative values are provided for limit or offset
func (q QueryParams) Validate() error {
	if q.Offset < 0 || q.Limit < 0 {
		return ErrInvalidPaginationQuery
	}
	return nil
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, filterAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls filter api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}

// GetFilter makes an authorised request to GET /filters
func (c *Client) GetFilter(ctx context.Context, input GetFilterInput) (*GetFilterResponse, error) {
	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, input.FilterID)
	clientlog.Do(ctx, "retrieving filter", service, uri)

	res, err := c.doGetWithAuthHeaders(
		ctx,
		input.UserAuthToken,
		input.ServiceAuthToken,
		input.CollectionID,
		uri,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	defer closeResponseBody(ctx, res)

	eTag, err := headers.GetResponseETag(res)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, errors.Wrap(err, "failed to get ETag header")
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	logData := log.Data{
		"uri":           uri,
		"response_body": string(b),
	}

	resp := GetFilterResponse{
		ETag: eTag,
	}

	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, dperrors.New(
			errors.Wrap(err, "failed to unmarshal response"),
			res.StatusCode,
			logData,
		)
	}

	if res.StatusCode != http.StatusOK {
		return nil, dperrors.New(
			errors.Errorf("error(s) returned by %s", uri),
			res.StatusCode,
			logData,
		)
	}

	return &resp, nil
}

// GetOutput returns a filter output job for a given filter output id, unmarshalled as a Model struct
func (c *Client) GetOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (m Model, err error) {
	b, err := c.GetOutputBytes(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(b, &m)
	return m, err
}

// GetOutputBytes returns a filter output job for a given filter output id as a byte array
func (c *Client) GetOutputBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) ([]byte, error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s", c.hcCli.URL, filterOutputID)
	clientlog.Do(ctx, "retrieving filter output", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, uri)
	if err != nil {
		return nil, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// UpdateFilterOutput performs a PUT operation to update the filter with the provided filterOutput model
func (c *Client) UpdateFilterOutput(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID string, model *Model) error {
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}

	return c.UpdateFilterOutputBytes(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID, b)
}

// UpdateFilterOutputBytes performs a PUT operation to update the filter with the provided byte array
func (c *Client) UpdateFilterOutputBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID string, b []byte) error {
	uri := fmt.Sprintf("%s/filter-outputs/%s", c.hcCli.URL, filterJobID)

	clientlog.Do(ctx, "updating filter output", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return fmt.Errorf("failed to set download service token: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}
	return nil
}

// AddEvent performs a POST operation to update the filter with the provided event
func (c *Client) AddEvent(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, filterJobID string, event *Event) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/filter-outputs/%s/events", c.hcCli.URL, filterJobID)

	clientlog.Do(ctx, "adding event to filter output", service, uri, log.Data{
		"method":   "POST",
		"filterID": filterJobID,
		"body":     string(b),
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return fmt.Errorf("failed to set download service token: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}
	return nil
}

// GetDimension returns information on a requested dimension name for a given filterID unmarshalled as a Dimension struct
func (c *Client) GetDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (dim Dimension, eTag string, err error) {
	b, eTag, err := c.GetDimensionBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name)
	if err != nil {
		return dim, "", err
	}

	err = json.Unmarshal(b, &dim)
	return dim, eTag, err
}

// GetDimensionBytes returns information on a requested dimension name for a given filterID as a byte array
func (c *Client) GetDimensionBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (body []byte, eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)
	clientlog.Do(ctx, "retrieving dimension information", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return nil, "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	body, err = ioutil.ReadAll(resp.Body)
	return body, eTag, err
}

// GetDimensions will return the dimensions associated with the provided filter id as an array of Dimension structs
func (c *Client) GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string, q *QueryParams) (dims Dimensions, eTag string, err error) {
	b, eTag, err := c.GetDimensionsBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, q)
	if err != nil {
		return dims, "", err
	}

	err = json.Unmarshal(b, &dims)
	return dims, eTag, err
}

// GetDimensionsBytes will return the dimensions associated with the provided filter id as a byte array
func (c *Client) GetDimensionsBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string, q *QueryParams) (body []byte, eTag string, err error) {

	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.hcCli.URL, filterID)
	if q != nil {
		if err := q.Validate(); err != nil {
			return nil, "", err
		}
		uri = fmt.Sprintf("%s?offset=%d&limit=%d", uri, q.Offset, q.Limit)
	}

	clientlog.Do(ctx, "retrieving all dimensions for given filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	body, err = ioutil.ReadAll(resp.Body)
	return body, eTag, err
}

// GetDimensionOptions retrieves a list of the dimension options unmarshalled as an array of DimensionOption structs
func (c *Client) GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *QueryParams) (opts DimensionOptions, eTag string, err error) {
	b, eTag, err := c.GetDimensionOptionsBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, q)
	if err != nil {
		return opts, "", err
	}

	err = json.Unmarshal(b, &opts)
	return opts, eTag, err
}

// GetDimensionOptionsBytes retrieves a list of the dimension options as a byte array
func (c *Client) GetDimensionOptionsBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q *QueryParams) (body []byte, eTag string, err error) {

	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.hcCli.URL, filterID, name)
	if q != nil {
		if err := q.Validate(); err != nil {
			return nil, "", err
		}
		uri = fmt.Sprintf("%s?offset=%d&limit=%d", uri, q.Offset, q.Limit)
	}
	clientlog.Do(ctx, "retrieving selected dimension options for filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return nil, "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	body, err = ioutil.ReadAll(resp.Body)
	return body, eTag, err
}

// GetDimensionOptionsInBatches retrieves a list of the dimension options in concurrent batches and accumulates the results.
// If the ETag changes from one batch to another, the process will be aborted and an ErrBatchETagMismatch error will be returned. You may retry the call in this case.
func (c *Client) GetDimensionOptionsInBatches(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, batchSize, maxWorkers int) (opts DimensionOptions, eTag string, err error) {

	// Function to aggregate items.
	// For the first received batch, as we have the total count information, will initialise the final structure of items with a fixed size equal to TotalCount.
	// This serves two purposes:
	//   - We can guarantee, even with concurrent calls, that values are returned in the same order that the API defines, by offsetting the index.
	//   - We do a single memory allocation for the final array, making the code more memory efficient.
	var processBatch DimensionOptionsBatchProcessor = func(b DimensionOptions, eTag string) (abort bool, err error) {
		if len(opts.Items) == 0 {
			opts.TotalCount = b.TotalCount
			opts.Items = make([]DimensionOption, b.TotalCount)
			opts.Count = b.TotalCount
		}
		for i := 0; i < len(b.Items); i++ {
			opts.Items[i+b.Offset] = b.Items[i]
		}
		return false, nil
	}

	// call filter API GetOptions in batches and aggregate the responses, enforcing ETag check
	eTag, err = c.GetDimensionOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, processBatch, batchSize, maxWorkers, true)
	if err != nil {
		return DimensionOptions{}, "", err
	}
	return opts, eTag, nil
}

// GetDimensionOptionsBatchProcess gets the filter options for a dimension from filter API in batches, and calls the provided function for each batch.
// If checkETag is true, then the ETag will be validated for each batch call. If it changes from one batch to another, an ErrBatchETagMismatch error will be returned.
// Unless your processBatch function performs some call to modify the same filter, it is recommended to set checkETag to true, and you may retry this call if it fails with ErrBatchETagMismatch
func (c *Client) GetDimensionOptionsBatchProcess(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, processBatch DimensionOptionsBatchProcessor, batchSize, maxWorkers int, checkETag bool) (eTag string, err error) {
	isFirstGet := true
	eTag = ""

	// for each batch, obtain the dimensions starting at the provided offset, with a batch size limit.
	// if any returned ETag is different from the previous one, an error is returned
	batchGetter := func(offset int) (interface{}, int, string, error) {
		b, newETag, err := c.GetDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, &QueryParams{Offset: offset, Limit: batchSize})
		if checkETag && newETag != eTag && !isFirstGet {
			return nil, 0, "", ErrBatchETagMismatch
		}
		eTag = newETag
		isFirstGet = false
		return b, b.TotalCount, newETag, err
	}

	// cast and process the batch according to the provided method
	batchProcessor := func(b interface{}, batchETag string) (abort bool, err error) {
		v, ok := b.(DimensionOptions)
		if !ok {
			return true, ErrBatchUnexpectedType
		}
		return processBatch(v, batchETag)
	}

	return eTag, batch.ProcessInConcurrentBatches(batchGetter, batchProcessor, batchSize, maxWorkers)
}

// DeleteDimensionOptions completely removes the options array from a given dimension
func (c *Client) DeleteDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (string, error) {
	logData := log.Data{
		"filter_id":      filterID,
		"dimension_name": name,
	}

	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.hcCli.URL, filterID, name)
	clientlog.Do(ctx, "deleting selected dimension options", service, uri)

	res, err := c.doDeleteWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return "", errors.Wrap(err, "failed to delete dimension options")
	}

	defer closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusNoContent {
		return "", dperrors.New(
			errors.Wrap(&ErrInvalidFilterAPIResponse{http.StatusNoContent, res.StatusCode, uri}, "unexpected response"),
			res.StatusCode,
			logData,
		)
	}

	eTag, err := headers.GetResponseETag(res)
	if err != nil && err != headers.ErrHeaderNotFound {
		dperrors.New(
			errors.Wrap(err, "no ETag header found"),
			res.StatusCode,
			logData,
		)
	}
	return eTag, err
}

// CreateFlexibleBlueprint creates a flexible filter blueprint and returns the associated filterID and eTag
func (c *Client) CreateFlexibleBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, dimensions []ModelDimension, population_type string) (filterID, eTag string, err error) {
	ver, err := strconv.Atoi(version)
	if err != nil {
		return "", "", err
	}

	cb := createFlexBlueprintRequest{
		Dimensions:     dimensions,
		Dataset:        Dataset{DatasetID: datasetID, Edition: edition, Version: ver},
		PopulationType: population_type,
	}

	reqBody, err := json.Marshal(cb)
	if err != nil {
		return "", "", err
	}

	respBody, eTag, err := c.postBlueprint(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, reqBody)
	if err != nil {
		return "", "", err
	}

	var respData createFlexBlueprintResponse
	if err = json.Unmarshal(respBody, &respData); err != nil {
		return "", "", err
	}

	return respData.FilterID, eTag, nil
}

// CreateFlexibleBlueprintCustom creates a flexible filter blueprint with the 'custom' flag set to true
// and returns the associated filterID and eTag
func (c *Client) CreateFlexibleBlueprintCustom(ctx context.Context, uAuthToken, svcAuthToken, dlServiceToken string, req CreateFlexBlueprintCustomRequest) (filterID, eTag string, err error) {
	r := struct {
		CreateFlexBlueprintCustomRequest
		Custom bool `json:"custom"`
	}{
		CreateFlexBlueprintCustomRequest: req,
		Custom:                           true,
	}

	reqBody, err := json.Marshal(r)
	if err != nil {
		return
	}

	respBody, eTag, err := c.postBlueprint(
		ctx,
		uAuthToken,
		svcAuthToken,
		dlServiceToken,
		req.CollectionID,
		req.Dataset.DatasetID,
		req.Dataset.Edition,
		strconv.Itoa(req.Dataset.Version),
		reqBody,
	)
	if err != nil {
		return
	}

	var resp createFlexBlueprintResponse
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return
	}

	filterID = resp.FilterID

	return
}

// CreateBlueprint creates a filter blueprint and returns the associated filterID and eTag
func (c *Client) CreateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, names []string) (filterID, eTag string, err error) {
	ver, err := strconv.Atoi(version)
	if err != nil {
		return "", "", err
	}

	dimensions := make([]ModelDimension, len(names))
	for i, name := range names {
		dimensions[i] = ModelDimension{Name: name}
	}

	cb := createBlueprint{
		Dimensions: dimensions,
		Dataset:    Dataset{DatasetID: datasetID, Edition: edition, Version: ver},
	}

	reqBody, err := json.Marshal(cb)
	if err != nil {
		return "", "", err
	}

	respBody, eTag, err := c.postBlueprint(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version, reqBody)
	if err != nil {
		return "", "", err
	}

	if err = json.Unmarshal(respBody, &cb); err != nil {
		return "", "", err
	}

	return cb.FilterID, eTag, nil
}

func (c *Client) CreateCustomFilter(ctx context.Context, userAuthToken, serviceAuthToken, populationType string) (filterID string, err error) {
	uri := c.hcCli.URL + "/custom/filters"

	clientlog.Do(ctx, "attempting to create custom filter ", service, uri, log.Data{
		"method":         "POST",
		"populationType": populationType,
	})

	body := struct {
		PopulationType string `json:"population_type"`
	}{
		PopulationType: populationType,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return "", ErrInvalidFilterAPIResponse{ExpectedCode: http.StatusCreated, ActualCode: resp.StatusCode, URI: uri}
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var fresp createFlexBlueprintResponse
	if err = json.Unmarshal(respBody, &fresp); err != nil {
		return
	}

	filterID = fresp.FilterID
	return
}

func (c *Client) postBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, datasetID, edition, version string, reqBody []byte) ([]byte, string, error) {
	uri := c.hcCli.URL + "/filters"

	clientlog.Do(ctx, "attempting to create filter blueprint", service, uri, log.Data{
		"method":    "POST",
		"datasetID": datasetID,
		"edition":   edition,
		"version":   version,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return nil, "", fmt.Errorf("failed to set download service token: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return nil, "", ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, eTag, nil
}

// UpdateBlueprint will update a blueprint with a given filter model, providing the required IfMatch value to be sure the update is done in the expected object
func (c *Client) UpdateBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID string, m Model, doSubmit bool, ifMatch string) (Model, string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return m, "", err
	}

	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, m.FilterID)

	if doSubmit {
		uri += "?submitted=true"
	}

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return m, "", err
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return m, "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return m, "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return m, "", fmt.Errorf("failed to set download service token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return m, "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return m, "", err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return m, "", ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return m, "", err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, "", err
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return m, "", err
	}

	return m, eTag, nil
}

// SubmitFilter function to submit the request to submit a filter for a cantabular dataset.
// Should POST to /filters/{filterid}/submit in dp-cantabular-filter-flex-api microservice.
func (c *Client) SubmitFilter(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, ifMatch string, sfr SubmitFilterRequest) (*SubmitFilterResponse, string, error) {
	b, err := json.Marshal(sfr)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not marshal submit filter request")
	}

	uri := fmt.Sprintf("%s/filters/%s/submit", c.hcCli.URL, sfr.FilterID)

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": http.MethodPost,
		"body":   string(b),
	})

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create a new POST request")
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, "", errors.Wrap(err, "failed to set auth token")
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, "", errors.Wrap(err, "failed to set service auth token")
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return nil, "", errors.Wrap(err, "failed to set download service token")
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return nil, "", errors.Wrap(err, "failed to set if match")
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to create submit request")
	}
	defer closeResponseBody(ctx, resp)

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", errors.Wrap(err, "no ETag header found")
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to read the response body")
	}

	if resp.StatusCode != http.StatusAccepted {
		return nil, "", dperrors.New(
			errors.Errorf("error(s) returned by %s", uri),
			resp.StatusCode,
			log.Data{"response_body": string(b)},
		)
	}

	var r *SubmitFilterResponse
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, "", errors.Wrap(err, "failed to unmarshal the response")
	}

	return r, eTag, nil
}

// UpdateFlexBlueprint will update a blueprint with a given filter model, providing the required IfMatch value to be sure the update is done in the expected object
func (c *Client) UpdateFlexBlueprint(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID string, m Model, doSubmit bool, populationType string, ifMatch string) (Model, string, error) {
	m.PopulationType = populationType

	b, err := json.Marshal(m)
	if err != nil {
		return m, "", err
	}

	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, m.FilterID)

	if doSubmit {
		uri += "?submitted=true"
	}

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": "PUT",
		"body":   string(b),
	})

	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(b))
	if err != nil {
		return m, "", err
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return m, "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return m, "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return m, "", fmt.Errorf("failed to set download service token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return m, "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return m, "", err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return m, "", ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return m, "", err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, "", err
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return m, "", err
	}

	return m, eTag, nil
}

// AddDimensionValue adds a particular value to a filter job for a given filterID
// and dimension name
func (c *Client) AddDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.hcCli.URL, filterID, name, value)

	clientlog.Do(ctx, "adding dimension option to filter job", service, uri, log.Data{
		"method": "POST",
		"value":  value,
	})

	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return "", &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, nil
}

// AddDimensionValues adds the provided values to a dimension option list. This is performed in batches of size up to batchSize
func (c *Client) AddDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, values []string, batchSize int, ifMatch string) (latestETag string, err error) {
	return c.PatchDimensionValues(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, values, []string{}, batchSize, ifMatch)
}

// RemoveDimensionValues removes the provided values from a dimension option list. This is performed with PATCH operations in batches of size up to batchSize.
func (c *Client) RemoveDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, values []string, batchSize int, ifMatch string) (latestETag string, err error) {
	return c.PatchDimensionValues(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, []string{}, values, batchSize, ifMatch)
}

// PatchDimensionValues adds and removes values from a dimension option list. If the same item is provided in the add and remove list, it will be removed. Duplicates in the same list will have no effect.
func (c *Client) PatchDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, addValues, removeValues []string, batchSize int, ifMatch string) (latestETag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)

	clientlog.Do(ctx, "attempting to patch a dimension options list in batches", service, uri, log.Data{
		"method":            http.MethodPatch,
		"collection_id":     collectionID,
		"filter_id":         filterID,
		"dimension_name":    name,
		"batch_size":        batchSize,
		"num_add_values":    len(addValues),
		"num_remove_values": len(removeValues),
	})

	// initialise latestETag to be ifMatch, in case no operation is performed
	latestETag = ifMatch

	// func to perform a provided PATCH call and handle errors and status code
	doPatchCall := func(patchBody []dprequest.Patch) error {
		resp, err := c.doPatchWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, ifMatch, patchBody)
		if err != nil {
			return err
		}
		defer closeResponseBody(ctx, resp)

		// check response code
		if resp.StatusCode != http.StatusOK {
			return &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}

		// get eTag from response
		latestETag, err = headers.GetResponseETag(resp)
		if err != nil && err != headers.ErrHeaderNotFound {
			return err
		}

		// ifMatch for next request is the eTag returned by the patch that has just been performed,
		// unless the caller specifically did not want eTgs validated
		if ifMatch != headers.IfMatchAnyETag {
			ifMatch = latestETag
		}

		return nil
	}

	if len(addValues)+len(removeValues) <= batchSize {

		// abort if no data is provided
		if len(addValues)+len(removeValues) == 0 {
			log.Info(ctx, "no PATCH operation has been sent because there aren't values to modify")
			return latestETag, nil
		}

		// we have less than a single batch size. We can do the add + remove operations in a single PATCH call
		patchBody := []dprequest.Patch{}
		if len(addValues) > 0 {
			patchBody = append(patchBody,
				dprequest.Patch{
					Op:    dprequest.OpAdd.String(),
					Path:  "/options/-",
					Value: addValues,
				})
		}
		if len(removeValues) > 0 {
			patchBody = append(patchBody,
				dprequest.Patch{
					Op:    dprequest.OpRemove.String(),
					Path:  "/options/-",
					Value: removeValues,
				})
		}

		if err := doPatchCall(patchBody); err != nil {
			log.Error(ctx, "error sending PATCH operation", err)
			return latestETag, err
		}

		log.Info(ctx, "successfully sent PATCH operation")
		return latestETag, nil
	}

	// func to perform an 'add' PATCH operation for a batch
	processAddPatch := func(items []string) error {
		patchBody := []dprequest.Patch{
			{
				Op:    dprequest.OpAdd.String(),
				Path:  "/options/-",
				Value: items,
			},
		}
		return doPatchCall(patchBody)
	}

	// func to perform a 'remove' PATCH operation for a batch
	processRemovePatch := func(items []string) error {
		patchBody := []dprequest.Patch{
			{
				Op:    dprequest.OpRemove.String(),
				Path:  "/options/-",
				Value: items,
			},
		}
		return doPatchCall(patchBody)
	}

	// perform sequential batched patches for add values
	numChunks, err := batch.ProcessInBatches(addValues, processAddPatch, batchSize)
	logData := log.Data{"num_successful_batches_added": numChunks}
	if err != nil {
		log.Error(ctx, "error sending PATCH operations in batches", err, logData)
		return latestETag, err
	}

	// perform sequential batched patches for remove values
	numChunks, err = batch.ProcessInBatches(removeValues, processRemovePatch, batchSize)
	logData["num_successful_batches_removed"] = numChunks
	if err != nil {
		log.Error(ctx, "error sending PATCH operations in batches", err, logData)
		return latestETag, err
	}

	log.Info(ctx, "successfully sent PATCH operations in batches", logData)
	return latestETag, nil
}

func (c *Client) UpdateDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, name, ifMatch string, dimension Dimension) (dim Dimension, eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, id, name)
	clientlog.Do(ctx, "updating filter dimension", service, uri, log.Data{
		"method":    http.MethodPut,
		"dimension": name,
	})

	reqBody, err := json.Marshal(dimension)
	if err != nil {
		return dimension, "", err
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(reqBody))
	if err != nil {
		return dimension, "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return dimension, "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return dimension, "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return dimension, "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return dimension, "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return dimension, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return dimension, "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return dimension, "", err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dimension, "", err
	}

	var updatedDimension Dimension

	if err = json.Unmarshal(respBody, &updatedDimension); err != nil {
		return updatedDimension, "", err
	}

	return updatedDimension, eTag, nil
}

// RemoveDimensionValue removes a particular value to a filter job for a given filterID and name
func (c *Client) RemoveDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.hcCli.URL, filterID, name, value)
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return "", err
	}

	clientlog.Do(ctx, "removing dimension option from filter job", service, uri, log.Data{
		"method": "DELETE",
		"value":  value,
	})

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusNoContent {
		return "", &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, nil
}

// RemoveDimension removes a given dimension from a filter job
func (c *Client) RemoveDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)

	clientlog.Do(ctx, "removing dimension from filter job", service, uri, log.Data{
		"method":    "DELETE",
		"dimension": "name",
	})

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusNoContent {
		err = &ErrInvalidFilterAPIResponse{http.StatusNoContent, resp.StatusCode, uri}
		return "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, err
}

// AddDimension adds a new dimension to a filter job
func (c *Client) AddDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, name, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, id, name)
	clientlog.Do(ctx, "adding dimension to filter job", service, uri, log.Data{
		"method":    "POST",
		"dimension": name,
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(`{}`))
	if err != nil {
		return "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		err = &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
		return "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, nil
}

// AddFlexDimension adds a new dimension to a filter job, with additional Cantabular-only fields
func (c *Client) AddFlexDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, name string, options []string, isAreaType bool, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.hcCli.URL, id)

	clientlog.Do(ctx, "adding dimension to filter job", service, uri, log.Data{
		"method":       "POST",
		"filter":       id,
		"dimension":    name,
		"options":      options,
		"is_area_type": isAreaType,
	})

	reqBody, err := json.Marshal(createFlexDimensionRequest{
		Name:       name,
		IsAreaType: isAreaType,
		Options:    options,
	})

	if err != nil {
		return "", fmt.Errorf("failed to marshal flex request body: %w", err)
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to make request to filter API: %w", err)
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}

	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}

	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to make filter request: %w", err)
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		err = &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
		return "", err
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", fmt.Errorf("unable to get reponse etag: %w", err)
	}

	return eTag, nil
}

// GetJobState will return the current state of the filter job unmarshalled as a Model struct
func (c *Client) GetJobState(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) (m Model, eTag string, err error) {
	b, eTag, err := c.GetJobStateBytes(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID)
	if err != nil {
		return m, "", err
	}

	err = json.Unmarshal(b, &m)
	return m, eTag, err
}

// GetJobStateBytes will return the current state of the filter job as a byte array
func (c *Client) GetJobStateBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) ([]byte, string, error) {
	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, filterID)
	clientlog.Do(ctx, "retrieving filter job state", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return nil, "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, "", err
	}

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, eTag, err
}

// SetDimensionValues creates or overwrites the options for a filter job dimension
func (c *Client) SetDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, options []string, ifMatch string) (eTag string, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)

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
		return "", err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return "", fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return "", fmt.Errorf("failed to set if match: %w", err)
	}

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return "", err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return "", &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	eTag, err = headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return "", err
	}

	return eTag, nil
}

// GetPreview attempts to retrieve a preview for a given filterOutputID unmarshalled as a Preview struct
func (c *Client) GetPreview(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) (p Preview, err error) {
	b, err := c.GetPreviewBytes(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID)
	if err != nil {
		return p, err
	}

	err = json.Unmarshal(b, &p)
	return p, err
}

// GetPreviewBytes attempts to retrieve a preview for a given filterOutputID as a byte array
func (c *Client) GetPreviewBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterOutputID string) ([]byte, error) {
	uri := fmt.Sprintf("%s/filter-outputs/%s/preview", c.hcCli.URL, filterOutputID)
	clientlog.Do(ctx, "retrieving preview for filter output job", service, uri, log.Data{
		"method":   "GET",
		"filterID": filterOutputID,
	})

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, uri)
	if err != nil {
		return nil, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	return ioutil.ReadAll(resp.Body)
}

// doGetWithAuthHeaders executes clienter.Do setting the user and service authentication token as a request header. Returns the http.Response and any error.
// It is the caller's responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set service auth token: %w", err)
	}
	return c.hcCli.Client.Do(ctx, req)
}

// doGetWithAuthHeadersAndWithDownloadToken executes clienter.Do setting the user and service authentication and download token as a request header. Returns the http.Response and any error.
// It is the caller's responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeadersAndWithDownloadToken(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set download service token: %w", err)
	}
	return c.hcCli.Client.Do(ctx, req)

} // doDeleteWithAuthHeadersAndWithDownloadToken executes clienter.Do setting the user and service authentication and download token as a request header.
// Returns the http.Response and any error.
// It is the caller's responsibility to ensure response.Body is closed on completion.
func (c *Client) doDeleteWithAuthHeadersAndWithDownloadToken(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set service auth token: %w", err)
	}
	return c.hcCli.Client.Do(ctx, req)
}

// doPatchWithAuthHeaders executes a PATCH request by using clienter.Do for the provided URI and patchBody.
// It sets the user and service authentication and coollectionID as a request header. Returns the http.Response and any error.
// It is the caller's responsibility to ensure response.Body is closed on completion.
func (c *Client) doPatchWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri, ifMatch string, patchBody []dprequest.Patch) (*http.Response, error) {

	// marshal the reuest body, as an array with the provided patch operation (http patch always accepts a list of patch operations)
	b, err := json.Marshal(patchBody)
	if err != nil {
		return nil, err
	}

	// create requets
	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	// set headers
	if err = headers.SetCollectionID(req, collectionID); err != nil {
		return nil, fmt.Errorf("failed to set collection id: %w", err)
	}
	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return nil, fmt.Errorf("failed to set if match: %w", err)
	}

	// do the request
	return c.hcCli.Client.Do(ctx, req)
}
