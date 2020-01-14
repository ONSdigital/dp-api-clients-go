package hierarchy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

const service = "hierarchy-api"

// ErrInvalidHierarchyAPIResponse is returned when the hierarchy api does not respond
// with a valid status
type ErrInvalidHierarchyAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidHierarchyAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from hierarchy api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from hierarchy api if an error is returned
func (e ErrInvalidHierarchyAPIResponse) Code() int {
	return e.actualCode
}

var _ error = ErrInvalidHierarchyAPIResponse{}

// Client is a hierarchy api client which can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
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

// New creates a new instance of Client with a given filter api url
func New(hierarchyAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: hierarchyAPIURL,
	}
}

// Checker calls hierarchy api health endpoint and returns a check object to the caller.
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

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	ctx := context.Background()
	endpoint := "/health"

	clientlog.Do(ctx, "checking health", service, endpoint)

	resp, err := c.cli.Get(ctx, c.url+endpoint)
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
		return service, &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, endpoint}
	}

	return service, nil
}

// GetRoot returns the root hierarchy response from the hierarchy API
func (c *Client) GetRoot(ctx context.Context, instanceID, name string) (Model, error) {
	path := fmt.Sprintf("/hierarchies/%s/%s", instanceID, name)

	clientlog.Do(ctx, "retrieving hierarchy", service, path, log.Data{
		"method":      "GET",
		"instance_id": instanceID,
		"dimension":   name,
	})

	return c.getHierarchy(ctx, path)
}

// GetChild returns a child of a given hierarchy and code
func (c *Client) GetChild(ctx context.Context, instanceID, name, code string) (Model, error) {
	path := fmt.Sprintf("/hierarchies/%s/%s/%s", instanceID, name, code)

	clientlog.Do(ctx, "retrieving hierarchy", service, path, log.Data{
		"method":      "GET",
		"instance_id": instanceID,
		"dimension":   name,
		"code":        code,
	})

	return c.getHierarchy(ctx, path)
}

func (c *Client) getHierarchy(ctx context.Context, path string) (Model, error) {
	var m Model
	req, err := http.NewRequest("GET", c.url+path, nil)
	if err != nil {
		return m, err
	}

	resp, err := c.cli.Do(ctx, req)
	if err != nil {
		return m, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return m, &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, path}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(b, &m)
	return m, err
}

func (c *Client) callHealthcheckEndpoint(ctx context.Context, service, endpoint string) (string, error) {
	clientlog.Do(ctx, "checking health", service, endpoint)
	resp, err := c.cli.Get(ctx, c.url+endpoint)
	if err != nil {
		return service, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidHierarchyAPIResponse{http.StatusOK, resp.StatusCode, endpoint}
	}

	return service, nil
}
