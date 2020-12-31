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
	"github.com/ONSdigital/dp-api-clients-go/common"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
)

const service = "filter-api"

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	ExpectedCode int
	ActualCode   int
	URI          string
}

// Config contains any configuration required to send requests to the filter api
type Config struct {
	InternalToken string
	FlorenceToken string
}

// DimensionOptionsBatchProcessor is the type corresponding to a batch processing function for filter DimensionOptions
type DimensionOptionsBatchProcessor func(DimensionOptions) (abort bool, err error)

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

// Validate validates tht no negative values are provided for limit or offset
func (q QueryParams) Validate() error {
	if q.Offset < 0 || q.Limit < 0 {
		return errors.New("negative offsets or limits are not allowed")
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

// CloseResponseBody closes the response body and logs an error if unsuccessful
func CloseResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
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

	defer CloseResponseBody(ctx, resp)

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

	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)
	headers.SetDownloadServiceToken(req, downloadServiceToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}
	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}
	return nil
}

// GetDimension returns information on a requested dimension name for a given filterID unmarshalled as a Dimension struct
func (c *Client) GetDimension(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) (dim Dimension, err error) {
	b, err := c.GetDimensionBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name)
	if err != nil {
		return dim, err
	}

	err = json.Unmarshal(b, &dim)
	return dim, err
}

// GetDimensionBytes returns information on a requested dimension name for a given filterID as a byte array
func (c *Client) GetDimensionBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string) ([]byte, error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)
	clientlog.Do(ctx, "retrieving dimension information", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// GetDimensions will return the dimensions associated with the provided filter id as an array of Dimension structs
func (c *Client) GetDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string, q QueryParams) (dims Dimensions, err error) {
	b, err := c.GetDimensionsBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, q)
	if err != nil {
		return dims, err
	}

	err = json.Unmarshal(b, &dims)
	return dims, err
}

// GetDimensionsBytes will return the dimensions associated with the provided filter id as a byte array
func (c *Client) GetDimensionsBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID string, q QueryParams) ([]byte, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s/filters/%s/dimensions?offset=%d&limit=%d", c.hcCli.URL, filterID, q.Offset, q.Limit)

	clientlog.Do(ctx, "retrieving all dimensions for given filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// GetDimensionOptions retrieves a list of the dimension options unmarshalled as an array of DimensionOption structs
func (c *Client) GetDimensionOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q QueryParams) (opts DimensionOptions, err error) {
	b, err := c.GetDimensionOptionsBytes(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, q)
	if err != nil {
		return opts, err
	}

	err = json.Unmarshal(b, &opts)
	return opts, err
}

// GetDimensionOptionsBytes retrieves a list of the dimension options as a byte array
func (c *Client) GetDimensionOptionsBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, q QueryParams) ([]byte, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options?offset=%d&limit=%d", c.hcCli.URL, filterID, name, q.Offset, q.Limit)
	clientlog.Do(ctx, "retrieving selected dimension options for filter job", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)

	if err != nil {
		return nil, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusNoContent {
			err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// GetDimensionOptionsInBatches retrieves a list of the dimension options in concurrent batches and accumulates the results
func (c *Client) GetDimensionOptionsInBatches(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, batchSize, maxWorkers int) (opts DimensionOptions, err error) {

	// function to aggregate items. The first received batch will initialise the structure with items with a fixed size
	// and items are aggregated in order even if batches are received in the wrong order (e.g. due to concurrent calls)
	var processBatch DimensionOptionsBatchProcessor = func(batch DimensionOptions) (abort bool, err error) {
		if len(opts.Items) == 0 {
			opts.TotalCount = batch.TotalCount
			opts.Items = make([]DimensionOption, batch.TotalCount)
			opts.Count = batch.TotalCount
		}
		for i := 0; i < len(batch.Items); i++ {
			opts.Items[i+batch.Offset] = batch.Items[i]
		}
		return false, nil
	}

	// call filter API GetOptions in batches and aggregate the responses
	if err := c.GetDimensionOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, processBatch, batchSize, maxWorkers); err != nil {
		return DimensionOptions{}, err
	}
	return opts, nil
}

// GetDimensionOptionsBatchProcess gets the filter options for a dimension from filter API in batches, and calls the provided function for each batch.
func (c *Client) GetDimensionOptionsBatchProcess(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, processBatch DimensionOptionsBatchProcessor, batchSize, maxWorkers int) (err error) {

	// for each batch, obtain the dimensions starting at the provided offset, with a batch size limit
	batchGetter := func(offset int) (interface{}, int, error) {
		batch, err := c.GetDimensionOptions(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, QueryParams{Offset: offset, Limit: batchSize})
		return batch, batch.TotalCount, err
	}

	// cast and process the batch according to the provided method
	batchProcessor := func(batch interface{}) (abort bool, err error) {
		v, ok := batch.(DimensionOptions)
		if !ok {
			return true, errors.New("wrong type")
		}
		return processBatch(v)
	}

	return common.ProcessInConcurrentBatches(batchGetter, batchProcessor, batchSize, maxWorkers)
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

	uri := c.hcCli.URL + "/filters"
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

	resp, err := c.hcCli.Client.Do(ctx, req)
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

	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, m.FilterID)

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

	resp, err := c.hcCli.Client.Do(ctx, req)
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
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.hcCli.URL, filterID, name, value)

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

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}
	return nil
}

// AddDimensionValues adds the provided values to a dimension option list. This is performed in batches of size up to batchSize
func (c *Client) AddDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, values []string, batchSize int) error {
	return c.PatchDimensionValues(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, values, []string{}, batchSize)
}

// RemoveDimensionValues removes the provided values from a dimension option list. This is performed with PATCH operations in batches of size up to batchSize.
func (c *Client) RemoveDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, values []string, batchSize int) error {
	return c.PatchDimensionValues(ctx, userAuthToken, serviceAuthToken, collectionID, filterID, name, []string{}, values, batchSize)
}

// PatchDimensionValues adds and removes values from a dimension option list. If the same item is provided in the add and remove list, it will be removed. Duplicates in the same list will have no effect.
func (c *Client) PatchDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, addValues, removeValues []string, batchSize int) error {
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

	// func to perform a provided PATCH call and handle errors and status code
	doPatchCall := func(patchBody []dprequest.Patch) error {
		resp, err := c.doPatchWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, patchBody)
		if err != nil {
			return err
		}
		defer CloseResponseBody(ctx, resp)

		// check response code
		if resp.StatusCode != http.StatusOK {
			return &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		}
		return nil
	}

	if len(addValues)+len(removeValues) <= batchSize {

		// abort if no data is provided
		if len(addValues)+len(removeValues) == 0 {
			log.Event(ctx, "no PATCH operation has been sent because there aren't values to modify", log.INFO)
			return nil
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
			log.Event(ctx, "error sending PATCH operation", log.ERROR, log.Error(err))
			return err
		}

		log.Event(ctx, "successfully sent PATCH operation", log.INFO)
		return nil
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

	// perform batched patches for add values
	numChunks, err := processInBatches(addValues, processAddPatch, batchSize)
	logData := log.Data{"num_successful_batches_added": numChunks}
	if err != nil {
		log.Event(ctx, "error sending PATCH operations in batches", log.ERROR, logData, log.Error(err))
		return err
	}

	// perform batched patches for remove values
	numChunks, err = processInBatches(removeValues, processRemovePatch, batchSize)
	logData["num_successful_batches_removed"] = numChunks
	if err != nil {
		log.Event(ctx, "error sending PATCH operations in batches", log.ERROR, logData, log.Error(err))
		return err
	}

	log.Event(ctx, "successfully sent PATCH operations in batches", log.INFO, logData)
	return nil
}

// RemoveDimensionValue removes a particular value to a filter job for a given filterID and name
func (c *Client) RemoveDimensionValue(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name, value string) error {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options/%s", c.hcCli.URL, filterID, name, value)
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

	resp, err := c.hcCli.Client.Do(ctx, req)
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
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, filterID, name)

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

	resp, err := c.hcCli.Client.Do(ctx, req)
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
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.hcCli.URL, id, name)
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

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return errors.New("invalid status from filter api")
	}

	return nil
}

// GetJobState will return the current state of the filter job unmarshalled as a Model struct
func (c *Client) GetJobState(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) (m Model, err error) {
	b, err := c.GetJobStateBytes(ctx, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(b, &m)
	return m, err
}

// GetJobStateBytes will return the current state of the filter job as a byte array
func (c *Client) GetJobStateBytes(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, collectionID, filterID string) ([]byte, error) {
	uri := fmt.Sprintf("%s/filters/%s", c.hcCli.URL, filterID)
	clientlog.Do(ctx, "retrieving filter job state", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri)
	if err != nil {
		return nil, err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// SetDimensionValues creates or overwrites the options for a filter job dimension
func (c *Client) SetDimensionValues(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, filterID, name string, options []string) error {
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
		return err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return err
	}

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusCreated {
		return &ErrInvalidFilterAPIResponse{http.StatusCreated, resp.StatusCode, uri}
	}

	return nil
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

	defer CloseResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	return ioutil.ReadAll(resp.Body)
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
	return c.hcCli.Client.Do(ctx, req)
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
	return c.hcCli.Client.Do(ctx, req)
}

// doPatchWithAuthHeaders executes a PATCH request by using clienter.Do for the provided URI and patchBody.
// It sets the user and service authentication and coollectionID as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doPatchWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, patchBody []dprequest.Patch) (*http.Response, error) {

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
	headers.SetCollectionID(req, collectionID)
	headers.SetUserAuthToken(req, userAuthToken)
	headers.SetServiceAuthToken(req, serviceAuthToken)

	// do the request
	return c.hcCli.Client.Do(ctx, req)
}
