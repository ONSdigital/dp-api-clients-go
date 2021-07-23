package health

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	// StatusMessage contains a map of messages to service response statuses
	StatusMessage = map[string]string{
		health.StatusOK:       " is ok",
		health.StatusWarning:  " is degraded, but at least partially functioning",
		health.StatusCritical: " functionality is unavailable or non-functioning",
	}
)

// ErrInvalidAppResponse is returned when an app does not respond
// with a valid status
type ErrInvalidAppResponse struct {
	ExpectedCode int
	ActualCode   int
	URI          string
}

// Client represents an app client
type Client struct {
	Client dphttp.Clienter
	URL    string
	Name   string
}

// NewClient creates a new instance of Client with a given app url
func NewClient(name, url string) *Client {
	return NewClientWithClienter(name, url, dphttp.NewClient())
}

// NewClientWithClienter creates a new instance of Client with a given app name and url, and the provided clienter
func NewClientWithClienter(name, url string, clienter dphttp.Clienter) *Client {
	c := &Client{
		Client: clienter,
		URL:    url,
		Name:   name,
	}

	// healthcheck client should not retry when calling a healthcheck endpoint,
	// append to current paths as to not change the client setup by service
	paths := c.Client.GetPathsWithNoRetries()
	paths = append(paths, "/health", "/healthcheck")
	c.Client.SetPathsWithNoRetries(paths)

	return c
}

// CreateCheckState creates a new check state object
func CreateCheckState(service string) (check health.CheckState) {
	check = *health.NewCheckState(service)

	return check
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAppResponse) Error() string {
	return fmt.Sprintf("invalid response from downstream service - should be: %d, got: %d, path: %s",
		e.ExpectedCode,
		e.ActualCode,
		e.URI,
	)
}

// Checker calls an app health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx context.Context, state *health.CheckState) error {
	service := c.Name
	logData := log.Data{
		"service": service,
	}

	code, err := c.get(ctx, "/health")
	// Apps may still have /healthcheck endpoint
	// instead of a /health one
	if code == http.StatusNotFound || code == http.StatusUnauthorized {
		code, err = c.get(ctx, "/healthcheck")
	}
	if err != nil {
		log.Error(ctx, "failed to request service health", err, logData)
	}

	switch code {
	case 0: // When there is a problem with the client return error in message
		state.Update(health.StatusCritical, err.Error(), 0)
	case 200:
		message := generateMessage(service, health.StatusOK)
		state.Update(health.StatusOK, message, code)
	case 429:
		message := generateMessage(service, health.StatusWarning)
		state.Update(health.StatusWarning, message, code)
	default:
		message := generateMessage(service, health.StatusCritical)
		state.Update(health.StatusCritical, message, code)
	}

	return nil
}

func (c *Client) get(ctx context.Context, path string) (int, error) {
	clientlog.Do(ctx, "retrieving service health", c.Name, c.URL)

	req, err := http.NewRequest("GET", c.URL+path, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return 0, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode < 200 || (resp.StatusCode > 399 && resp.StatusCode != 429) {
		return resp.StatusCode, ErrInvalidAppResponse{http.StatusOK, resp.StatusCode, req.URL.Path}
	}

	return resp.StatusCode, nil
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body == nil {
		return
	}

	if err := resp.Body.Close(); err != nil {
		log.Error(ctx, "error closing http response body", err)
	}
}

func generateMessage(service string, state string) string {
	return service + StatusMessage[state]
}
