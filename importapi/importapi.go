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

	"github.com/ONSdigital/dp-api-clients-go/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

const service = "import-api"

// Client is an import api client which can be used to make requests to the API
type Client struct {
	cli rchttp.Clienter
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
func (c *Client) GetImportJob(ctx context.Context, importJobID, serviceToken string) (ImportJob, bool, error) {
	var importJob ImportJob
	path := c.url + "/jobs/" + importJobID

	jsonBody, httpCode, err := c.getJSON(ctx, path, serviceToken, 0, nil)
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
		log.Event(ctx, "GetImportJob", logData, log.Error(err))
		return importJob, isFatal, err
	}

	if err := json.Unmarshal(jsonBody, &importJob); err != nil {
		log.Event(ctx, "GetImportJob unmarshal", logData, log.Error(err))
		return ImportJob{}, true, err
	}

	return importJob, false, nil
}

// UpdateImportJobState tells the Import API that the state has changed of an Import job
func (c *Client) UpdateImportJobState(ctx context.Context, jobID, serviceToken string, newState string) error {
	path := c.url + "/jobs/" + jobID
	jsonUpload := []byte(`{"state":"` + newState + `"}`)

	jsonResult, httpCode, err := c.putJSON(ctx, path, serviceToken, 0, jsonUpload)
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
		log.Event(ctx, "UpdateImportJobState", logData, log.Error(err))
		return err
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path, serviceToken string, attempts int, vars url.Values) ([]byte, int, error) {
	return callJSONAPI(ctx, c.cli, "GET", path, serviceToken, vars)
}

func (c *Client) putJSON(ctx context.Context, path, serviceToken string, attempts int, payload []byte) ([]byte, int, error) {
	return callJSONAPI(ctx, c.cli, "PUT", path, serviceToken, payload)
}

func callJSONAPI(ctx context.Context, client rchttp.Clienter, method, path, serviceToken string, payload interface{}) ([]byte, int, error) {

	logData := log.Data{"url": path, "method": method}

	URL, err := url.Parse(path)
	if err != nil {
		log.Event(ctx, "Failed to create url for API call", logData, log.Error(err))
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
		log.Event(ctx, "Failed to create request for API", logData, log.Error(err))
		return nil, 0, err
	}

	// add a service token to request where one has been provided
	headers.SetServiceAuthToken(req, serviceToken)

	resp, err := client.Do(ctx, req)
	if err != nil {
		log.Event(ctx, "Failed to action API", logData, log.Error(err))
		return nil, 0, err
	}
	defer closeResponseBody(ctx, resp)

	logData["httpCode"] = resp.StatusCode
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		log.Event(ctx, "unexpected status code from API", logData)
	}

	jsonBody, err := getBody(resp)
	if err != nil {
		log.Event(ctx, "Failed to read body from API", logData, log.Error(err))
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

	return b, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}
