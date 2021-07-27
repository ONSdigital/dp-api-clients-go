package importapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
)

const service = "import-api"

// Client is an import api client which can be used to make requests to the API
type Client struct {
	cli dphttp.Clienter
	url string
}

// New creates new instance of Client with a give import api url
func New(importAPIURL string) *Client {
	hcClient := healthcheck.NewClient(service, importAPIURL)

	return &Client{
		cli: hcClient.Client,
		url: importAPIURL,
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

// ImportJob comes from the Import API and links an import job to its (other) instances
type ImportJob struct {
	JobID     string               `json:"id"`
	Links     LinkMap              `json:"links,omitempty"`
	Processed []ProcessedInstances `json:"processed_instances,omitempty"`
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

// ProcessedInstances holds the ID and the number of code lists that have been processed during an import process for an instance
type ProcessedInstances struct {
	ID             string `json:"id,omitempty"`
	RequiredCount  int    `json:"required_count,omitempty"`
	ProcessedCount int    `json:"processed_count,omitempty"`
}

// stateData represents a json with a single state filed
type stateData struct {
	State string `json:"state"`
}

// Checker calls import api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	hcClient := healthcheck.Client{
		Client: c.cli,
		URL:    c.url,
		Name:   service,
	}

	return hcClient.Checker(ctx, check)
}

// GetImportJob asks the Import API for the details for an Import job
func (c *Client) GetImportJob(ctx context.Context, importJobID, serviceToken string) (importJob ImportJob, err error) {
	uri := fmt.Sprintf("%s/jobs/%s", c.url, importJobID)

	resp, err := c.doGet(ctx, uri, serviceToken, 0, nil)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	jsonBody, err := getBody(resp)
	if err != nil {
		log.Event(ctx, "Failed to read body from API", log.ERROR, log.Error(err))
		return importJob, err
	}

	logData := log.Data{
		"uri":         uri,
		"importJobID": importJobID,
		"httpCode":    resp.StatusCode,
		"jsonBody":    string(jsonBody),
	}

	if resp.StatusCode != http.StatusOK {
		return importJob, NewAPIResponse(resp, uri)
	}

	if err := json.Unmarshal(jsonBody, &importJob); err != nil {
		log.Event(ctx, "GetImportJob unmarshal", log.ERROR, logData, log.Error(err))
		return importJob, err
	}

	return importJob, nil
}

// UpdateImportJobState tells the Import API that the state has changed of an Import job
func (c *Client) UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState string) error {
	uri := fmt.Sprintf("%s/jobs/%s", c.url, jobID)

	jsonUpload, err := json.Marshal(&stateData{newState})
	if err != nil {
		return err
	}

	logData := log.Data{
		"uri":         uri,
		"importJobID": jobID,
		"newState":    newState,
	}

	resp, err := c.doPut(ctx, uri, serviceToken, 0, jsonUpload)
	if err != nil {
		log.Event(ctx, "UpdateImportJobState", log.ERROR, logData, log.Error(err))
		return err
	}
	defer closeResponseBody(ctx, resp)
	logData["httpCode"] = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		return NewAPIResponse(resp, uri)
	}
	return nil
}

func (c *Client) IncreaseProcessedInstanceCount(ctx context.Context, jobID, serviceToken, instanceID string) (procInst []ProcessedInstances, err error) {
	uri := fmt.Sprintf("%s/jobs/%s/processed/%s", c.url, jobID, instanceID)

	logData := log.Data{
		"uri":         uri,
		"job_id":      jobID,
		"instance_id": instanceID,
	}

	resp, err := c.doPut(ctx, uri, serviceToken, 0, nil)
	if err != nil {
		log.Event(ctx, "error increaseing the instance count in import api", log.ERROR, logData, log.Error(err))
		return nil, err
	}
	defer closeResponseBody(ctx, resp)
	logData["httpCode"] = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		return nil, NewAPIResponse(resp, uri)
	}

	jsonBody, err := getBody(resp)
	if err != nil {
		log.Event(ctx, "failed to read body from api response", log.ERROR, log.Error(err))
		return nil, err
	}

	if err := json.Unmarshal(jsonBody, &procInst); err != nil {
		log.Event(ctx, "failed to unmarshal api response body", log.ERROR, logData, log.Error(err))
		return nil, err
	}

	return procInst, nil
}

func (c *Client) doGet(ctx context.Context, uri, serviceToken string, attempts int, vars url.Values) (*http.Response, error) {
	return doCall(ctx, c.cli, "GET", uri, serviceToken, vars)
}

func (c *Client) doPut(ctx context.Context, uri, serviceToken string, attempts int, payload []byte) (*http.Response, error) {
	return doCall(ctx, c.cli, "PUT", uri, serviceToken, payload)
}

func doCall(ctx context.Context, client dphttp.Clienter, method, uri, serviceToken string, payload interface{}) (*http.Response, error) {

	logData := log.Data{"uri": uri, "method": method}

	URL, err := url.Parse(uri)
	if err != nil {
		log.Event(ctx, "Failed to create url for API call", log.ERROR, logData, log.Error(err))
		return nil, err
	}
	uri = URL.String()
	logData["url"] = uri

	var req *http.Request

	if payload != nil && method != http.MethodGet {
		req, err = http.NewRequest(method, uri, bytes.NewReader(payload.([]byte)))
		req.Header.Add("Content-type", "application/json")
		logData["payload"] = string(payload.([]byte))
	} else {
		req, err = http.NewRequest(method, uri, nil)

		if payload != nil && method == http.MethodGet {
			req.URL.RawQuery = payload.(url.Values).Encode()
			logData["payload"] = payload.(url.Values)
		}
	}
	// check above req had no errors
	if err != nil {
		log.Event(ctx, "Failed to create request for API", log.ERROR, logData, log.Error(err))
		return nil, err
	}

	// add a service token to request where one has been provided
	dprequest.AddServiceTokenHeader(req, serviceToken)

	resp, err := client.Do(ctx, req)
	if err != nil {
		log.Event(ctx, "Failed to action API", log.ERROR, logData, log.Error(err))
		return nil, err
	}

	return resp, nil
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

	return b, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}
