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
	check *health.Check
	cli   rchttp.Clienter
	url   string
}

// CloseResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.Error(err))
	}
}

// New creates a new instance of Client with a given hierarchy api url
func New(hierarchyAPIURL string) *Client {
	hcClient := healthcheck.NewClient(service, hierarchyAPIURL)

	return &Client{
		check: hcClient.Check,
		cli:   hcClient.Client,
		url:   hierarchyAPIURL,
	}
}

// Checker calls hierarchy api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	hcClient := healthcheck.Client{
		Check:  c.check,
		Client: c.cli,
	}

	return hcClient.Checker(ctx)
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
