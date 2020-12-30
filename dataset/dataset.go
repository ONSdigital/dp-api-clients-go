package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
)

const service = "dataset-api"

const maxIDs = 200

// MaxIDs returns the maximum number of IDs acceptable in a list
var MaxIDs = func() int {
	return maxIDs
}

// State - iota enum of possible states
type State int

// Possible values for a State of the resource. It can only be one of the following:
// TODO these states should be enforced in all the 'POST' and 'PUT' operations that can modify states of resources
const (
	StateCreated State = iota
	StateSubmitted
	StateCompleted        // Instances only
	StateFailed           // Instances only
	StateEditionConfirmed // instances and versions only
	StateAssociated       // not editions
	StatePublished
)

var stateValues = []string{"created", "submitted", "completed", "failed", "edition-confirmed", "associated", "published"}

// String returns the string representation of a state
func (s State) String() string {
	return stateValues[s]
}

// ErrInvalidDatasetAPIResponse is returned when the dataset api does not respond
// with a valid status
type ErrInvalidDatasetAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidDatasetAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from dataset api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from dataset api if an error is returned
func (e ErrInvalidDatasetAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidDatasetAPIResponse{}

// Client is a dataset api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// QueryParams represents the possible query parameters that a caller can provide
type QueryParams struct {
	Offset int
	Limit  int
	IDs    []string
}

// Validate validates tht no negative values are provided for limit or offset, and that the length of IDs is lower than the maximum
func (q QueryParams) Validate() error {
	if q.Offset < 0 || q.Limit < 0 {
		return errors.New("negative offsets or limits are not allowed")
	}
	if len(q.IDs) > MaxIDs() {
		return fmt.Errorf("too many query parameters have been provided. Maximum allowed: %d", MaxIDs())
	}
	return nil
}

// NewAPIClient creates a new instance of Client with a given dataset api url and the relevant tokens
func NewAPIClient(datasetAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, datasetAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// NewAPIClientWithMaxRetries creates a new instance of Client with a given dataset api url and the relevant tokens,
// setting a number of max retires for the HTTP client
func NewAPIClientWithMaxRetries(datasetAPIURL string, maxRetries int) *Client {
	hcClient := healthcheck.NewClient(service, datasetAPIURL)
	if maxRetries > 0 {
		hcClient.Client.SetMaxRetries(maxRetries)
	}

	return &Client{
		hcClient,
	}
}

// Checker calls dataset api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// Get returns dataset level information for a given dataset id
func (c *Client) Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m DatasetDetails, err error) {
	uri := fmt.Sprintf("%s/datasets/%s", c.hcCli.URL, datasetID)

	clientlog.Do(ctx, "retrieving dataset", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && (serviceAuthToken != "" || userAuthToken != "") {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetByPath returns dataset level information for a given dataset path
func (c *Client) GetByPath(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, path string) (m DatasetDetails, err error) {
	uri := fmt.Sprintf("%s/%s", c.hcCli.URL, strings.Trim(path, "/"))

	clientlog.Do(ctx, "retrieving data from dataset API", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && (serviceAuthToken != "" || userAuthToken != "") {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetDatasets returns the list of datasets
func (c *Client) GetDatasets(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string) (m List, err error) {
	uri := fmt.Sprintf("%s/datasets", c.hcCli.URL)

	clientlog.Do(ctx, "retrieving datasets", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
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

// PutDataset update the dataset
func (c *Client) PutDataset(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string, d DatasetDetails) error {
	uri := fmt.Sprintf("%s/datasets/%s", c.hcCli.URL, datasetID)

	clientlog.Do(ctx, "updating dataset", service, uri)

	payload, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall dataset")
	}

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// GetEdition retrieves a single edition document from a given datasetID and edition label
func (c *Client) GetEdition(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition string) (m Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s", c.hcCli.URL, datasetID, edition)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	if next, ok := body["next"]; ok && userAuthToken != "" {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetEditions returns all editions for a dataset
func (c *Client) GetEditions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m []Edition, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions", c.hcCli.URL, datasetID)

	clientlog.Do(ctx, "retrieving dataset editions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return nil, nil
	}

	if _, ok := body["items"].([]interface{})[0].(map[string]interface{})["next"]; ok && userAuthToken != "" {
		var items []map[string]interface{}
		for _, item := range body["items"].([]interface{}) {
			items = append(items, item.(map[string]interface{})["next"].(map[string]interface{}))
		}
		parentItems := make(map[string]interface{})
		parentItems["items"] = items
		b, err = json.Marshal(parentItems)
		if err != nil {
			return
		}
	}

	editions := struct {
		Items []Edition `json:"items"`
	}{}
	err = json.Unmarshal(b, &editions)
	m = editions.Items
	return
}

// GetVersions gets all versions for an edition from the dataset api
func (c *Client) GetVersions(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition string) (m []Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions", c.hcCli.URL, datasetID, edition)

	clientlog.Do(ctx, "retrieving dataset versions", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	versions := struct {
		Items []Version `json:"items"`
	}{}

	err = json.Unmarshal(b, &versions)
	m = versions.Items
	return
}

// GetVersion gets a specific version for an edition from the dataset api
func (c *Client) GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m Version, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.hcCli.URL, datasetID, edition, version)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	resp, err := c.doGetWithAuthHeadersAndWithDownloadToken(ctx, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, uri)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetInstance returns an instance from the dataset api
func (c *Client) GetInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) (m Instance, err error) {
	b, err := c.GetInstanceBytes(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetInstanceBytes returns an instance as bytes from the dataset api
func (c *Client) GetInstanceBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) (b []byte, err error) {
	uri := fmt.Sprintf("%s/instances/%s", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

// GetInstanceDimensionsBytes returns a list of dimensions for an instance as bytes from the dataset api
func (c *Client) GetInstanceDimensionsBytes(ctx context.Context, userAuthToken, serviceAuthToken, instanceID string) (b []byte, err error) {
	uri := fmt.Sprintf("%s/instances/%s/dimensions", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "retrieving instance dimensions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, "", uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

// GetInstances returns a list of all instances filtered by vars
func (c *Client) GetInstances(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, vars url.Values) (m Instances, err error) {
	uri := fmt.Sprintf("%s/instances", c.hcCli.URL)

	clientlog.Do(ctx, "retrieving dataset version", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, vars)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	json.Unmarshal(b, &m)
	return
}

// PutInstance updates an instance
func (c *Client) PutInstance(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string, i Instance) error {
	uri := fmt.Sprintf("%s/instances/%s", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "updating dataset version", service, uri)

	payload, err := json.Marshal(i)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall instance")
	}

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// PutInstanceState performs a PUT '/instances/<id>' with the string representation of the provided state
func (c *Client) PutInstanceState(ctx context.Context, serviceAuthToken, instanceID string, state State) error {
	payload, err := json.Marshal(stateData{State: state.String()})
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/instances/%s", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "putting state to instance", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, payload)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// PutInstanceData executes a put request to update instance data via the dataset API.
func (c *Client) PutInstanceData(ctx context.Context, serviceAuthToken, instanceID string, data JobInstance) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/instances/%s", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "putting data to instance", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, payload)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// PutInstanceImportTasks marks the import observation task state for an instance
func (c *Client) PutInstanceImportTasks(ctx context.Context, serviceAuthToken, instanceID string, data InstanceImportTasks) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/instances/%s/import_tasks", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "updating instance import_tasks", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, payload)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// UpdateInstanceWithNewInserts increments the observation inserted count for an instance
func (c *Client) UpdateInstanceWithNewInserts(ctx context.Context, serviceAuthToken, instanceID string, observationsInserted int32) error {
	uri := fmt.Sprintf("%s/instances/%s/inserted_observations/%d", c.hcCli.URL, instanceID, observationsInserted)

	clientlog.Do(ctx, "updating instance inserted observations", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, nil)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// GetInstanceDimensions performs a 'GET /instances/<id>/dimensions' and returns the marshalled Dimensions struct
func (c *Client) GetInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string) (m Dimensions, err error) {
	uri := fmt.Sprintf("%s/instances/%s/dimensions", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "retrieving instance dimensions", service, uri)

	b, err := c.GetInstanceDimensionsBytes(ctx, "", serviceAuthToken, instanceID)
	if err != nil {
		return
	}

	json.Unmarshal(b, &m)
	return
}

// PostInstanceDimensions performs a 'POST /instances/<id>/dimensions' with the provided OptionPost
func (c *Client) PostInstanceDimensions(ctx context.Context, serviceAuthToken, instanceID string, data OptionPost) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/instances/%s/dimensions", c.hcCli.URL, instanceID)

	clientlog.Do(ctx, "posting options to instance dimensions", service, uri)

	resp, err := c.doPostWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, payload)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// PutInstanceDimensionOptionNodeID performs a 'PUT /instances/<id>/dimensions/<id>/options/<id>/node_id/<id>' to update the node_id of the specified dimension
func (c *Client) PutInstanceDimensionOptionNodeID(ctx context.Context, serviceAuthToken, instanceID, dimensionID, optionID, nodeID string) error {
	uri := fmt.Sprintf("%s/instances/%s/dimensions/%s/options/%s/node_id/%s", c.hcCli.URL, instanceID, dimensionID, optionID, nodeID)

	clientlog.Do(ctx, "updating instance dimension option node_id", service, uri)

	resp, err := c.doPutWithAuthHeaders(ctx, "", serviceAuthToken, "", uri, nil)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return NewDatasetAPIResponse(resp, uri)
	}
	return nil
}

// PutVersion update the version
func (c *Client) PutVersion(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID, edition, version string, v Version) error {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s", c.hcCli.URL, datasetID, edition, version)

	clientlog.Do(ctx, "updating version", service, uri)

	payload, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	resp, err := c.doPutWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, payload)
	if err != nil {
		return errors.Wrap(err, "http client returned error while attempting to make request")
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("incorrect http status, expected: 200, actual: %d, uri: %s", resp.StatusCode, uri)
	}
	return nil
}

// GetMetadataURL returns the URL for the metadata of a given dataset id, edition and version
func (c *Client) GetMetadataURL(id, edition, version string) string {
	return fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/metadata", c.hcCli.URL, id, edition, version)
}

// GetVersionMetadata returns the metadata for a given dataset id, edition and version
func (c *Client) GetVersionMetadata(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m Metadata, err error) {
	uri := c.GetMetadataURL(id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version metadata", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// GetVersionDimensions will return a list of dimensions for a given version of a dataset
func (c *Client) GetVersionDimensions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version string) (m VersionDimensions, err error) {
	uri := fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions", c.hcCli.URL, id, edition, version)

	clientlog.Do(ctx, "retrieving dataset version dimensions", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if err = json.Unmarshal(b, &m); err != nil {
		return
	}

	sort.Sort(m.Items)

	return
}

// GetOptions will return the options for a dimension
func (c *Client) GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q QueryParams) (m Options, err error) {
	if err := q.Validate(); err != nil {
		return Options{}, err
	}

	var uri string
	if len(q.IDs) > 0 {
		uri = fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options?id=%s", c.hcCli.URL, id, edition, version, dimension, strings.Join(q.IDs, ","))
	} else {
		uri = fmt.Sprintf("%s/datasets/%s/editions/%s/versions/%s/dimensions/%s/options?offset=%d&limit=%d", c.hcCli.URL, id, edition, version, dimension, q.Offset, q.Limit)
	}

	clientlog.Do(ctx, "retrieving options for dimension", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewDatasetAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &m)
	return
}

// NewDatasetAPIResponse creates an error response, optionally adding body to e when status is 404
func NewDatasetAPIResponse(resp *http.Response, uri string) (e *ErrInvalidDatasetAPIResponse) {
	e = &ErrInvalidDatasetAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read DatasetAPI body"
			return
		}
		defer closeResponseBody(nil, resp)

		e.body = string(b)
	}
	return
}

func addCollectionIDHeader(r *http.Request, collectionID string) {
	if len(collectionID) > 0 {
		r.Header.Add(dprequest.CollectionIDHeaderKey, collectionID)
	}
}

// doGetWithAuthHeaders executes clienter.Do setting the user and service authentication token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
// If url.Values are provided, they will be added as query parameters in the URL.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, values url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if values != nil {
		req.URL.RawQuery = values.Encode()
	}

	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

func (c *Client) doPostWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

func (c *Client) doPutWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// doGetWithAuthHeadersAndWithDownloadToken executes clienter.Do setting the user and service authentication and download token token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeadersAndWithDownloadToken(ctx context.Context, userAuthToken, serviceAuthToken, downloadserviceAuthToken, collectionID, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	dprequest.AddDownloadServiceTokenHeader(req, downloadserviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// closeResponseBody closes the response body and logs an error containing the context if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}
