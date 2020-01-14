package importapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
)

const service = "import-api"

// Client is an import api client which can be used to make requests to the API
type Client struct {
	client rchttp.Clienter
	url    string
}

// NewAPIClient creates a new API Client
func NewAPIClient(client rchttp.Clienter, apiURL string) *Client {
	return &Client{
		client: client,
		url:    apiURL,
	}
}

// ErrInvalidAPIResponse is returned when the api does not respond with a valid status
type ErrInvalidAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAPIResponse) Error() string {
	return fmt.Sprintf(
		"invalid response: %d from %s: %s, body: %s",
		e.actualCode,
		service,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from the api if an error is returned
func (e ErrInvalidAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidAPIResponse{}

// Checker calls hierarchy api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	hcClient := healthcheck.Client{
		Client: c.client,
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

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()
	endpoint := "/health"

	clientlog.Do(ctx, "checking health", service, endpoint)

	resp, err := c.client.Get(ctx, c.url+endpoint)
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	// Apps may still have /healthcheck endpoint instead of a /health one.
	if resp.StatusCode == http.StatusNotFound {
		endpoint = "/healthcheck"
		return c.callHealthcheckEndpoint(ctx, service, endpoint)
	}

	if resp.StatusCode != http.StatusOK {
		return service, NewAPIResponse(resp, "/healthcheck")
	}

	return service, nil
}

// ImportJob comes from the Import API and links an import job to its (other) instances
type ImportJob struct {
	JobID string  `json:"id"`
	Links LinkMap `json:"links,omitempty"`
}

// LinkMap is an array of instance links associated with am import job
type LinkMap struct {
	Instances []InstanceLink `json:"instances"`
}

// InstanceLink identifies an (instance or import-job) by id and url (from Import API)
type InstanceLink struct {
	ID   string `json:"id"`
	Link string `json:"href"`
}

// GetImportJob asks the Import API for the details for an Import job
func (api *Client) GetImportJob(ctx context.Context, importJobID, serviceToken string) (ImportJob, bool, error) {
	var importJob ImportJob
	path := api.url + "/jobs/" + importJobID

	jsonBody, httpCode, err := api.getJSON(ctx, path, serviceToken, 0, nil)
	if httpCode == http.StatusNotFound {
		return importJob, false, nil
	}
	logData := log.Data{
		"path":        path,
		"importJobID": importJobID,
		"httpCode":    httpCode,
		"jsonBody":    string(jsonBody),
	}
	var isFatal bool
	if err == nil && httpCode != http.StatusOK {
		if httpCode < http.StatusInternalServerError {
			isFatal = true
		}
		err = errors.New("Bad response while getting import job")
	} else {
		isFatal = true
	}
	if err != nil {
		log.Event(ctx, "GetImportJob", log.Error(err), logData)
		return importJob, isFatal, err
	}

	if err := json.Unmarshal(jsonBody, &importJob); err != nil {
		log.Event(ctx, "GetImportJob unmarshal", log.Error(err), logData)
		return ImportJob{}, true, err
	}

	return importJob, false, nil
}

// UpdateImportJobState tells the Import API that the state has changed of an Import job
func (api *Client) UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState string) error {
	path := api.url + "/jobs/" + jobID
	jsonUpload := []byte(`{"state":"` + newState + `"}`)

	jsonResult, httpCode, err := api.putJSON(ctx, path, serviceToken, 0, jsonUpload)
	logData := log.Data{
		"path":        path,
		"importJobID": jobID,
		"jsonUpload":  jsonUpload,
		"httpCode":    httpCode,
		"jsonResult":  jsonResult,
	}
	if err == nil && httpCode != http.StatusOK {
		err = errors.New("Bad HTTP response")
	}
	if err != nil {
		log.Event(ctx, "UpdateImportJobState", log.Error(err), logData)
		return err
	}
	return nil
}

func (api *Client) getJSON(ctx context.Context, path, serviceToken string, attempts int, vars url.Values) ([]byte, int, error) {
	return callJSONAPI(ctx, api.client, "GET", path, serviceToken, vars)
}

func (api *Client) putJSON(ctx context.Context, path, serviceToken string, attempts int, payload []byte) ([]byte, int, error) {
	return callJSONAPI(ctx, api.client, "PUT", path, serviceToken, payload)
}

func callJSONAPI(ctx context.Context, client rchttp.Clienter, method, path, serviceToken string, payload interface{}) ([]byte, int, error) {

	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.Event(ctx, "Failed to create url for API call", log.Error(err), logData)
		return nil, 0, err
	}
	path = URL.String()
	logData["url"] = path

	var req *http.Request

	if payload != nil && method != "GET" {
		req, err = http.NewRequest(method, path, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, path, nil)

		if payload != nil && method == "GET" {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}
	// check above req had no errors
	if err != nil {
		log.Event(ctx, "Failed to create request for API", log.Error(err), logData)
		return nil, 0, err
	}

	// add a service token to request where one has been provided
	common.AddServiceTokenHeader(req, serviceToken)

	resp, err := client.Do(ctx, req)
	if err != nil {
		log.Event(ctx, "Failed to action API", log.Error(err), logData)
		return nil, 0, err
	}

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Event(ctx, "unexpected status code from API", log.Data{"url": path, "method": method, "severity": 3})
	}

	jsonBody, err := getBody(resp)
	if err != nil {
		log.Event(ctx, "Failed to read body from API", log.Error(err), logData)
		return nil, resp.StatusCode, err
	}
	return jsonBody, resp.StatusCode, nil
}

// NewAPIResponse creates an error response, optionally adding body to e when status is 404
func NewAPIResponse(resp *http.Response, uri string) (e *ErrInvalidAPIResponse) {
	e = &ErrInvalidAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		body, err := getBody(resp)
		if err != nil {
			e.body = "Client failed to read response body"
			return
		}
		e.body = string(body)
	}
	return
}

func getBody(resp *http.Response) ([]byte, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = resp.Body.Close(); err != nil {
		log.Event(ctx, "closing body", log.Error(err))
		return nil, err
	}
	return b, nil
}

// CloseResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}

func (c *Client) callHealthcheckEndpoint(ctx context.Context, service, endpoint string) (string, error) {
	clientlog.Do(ctx, "checking health", service, endpoint)
	resp, err := c.client.Get(ctx, c.url+endpoint)
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return service, NewAPIResponse(resp, endpoint)
	}

	return service, nil
}
