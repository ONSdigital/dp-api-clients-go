package health

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

// ErrInvalidAPIResponse is returned when an api does not respond
// with a valid status.
type ErrInvalidAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Client represents an api client.
type Client struct {
	Client rchttp.Clienter
	Name   string
	URL    string
}

// NewClient creates a new instance of Client with a given api url.
func NewClient(name, url string) *Client {
	client := rchttp.NewClient()

	c := &Client{
		Client: client,
		Name:   name,
		URL:    url,
	}

	// Overwrite the default number of max retries on the new healthcheck client.
	c.Client.SetMaxRetries(0)

	return c
}

// Error should be called by the user to print out the stringified version of the error.
func (e ErrInvalidAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from downstream service - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Checker calls an api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	errorMessage := ""

	logData := log.Data{
		"api": c.Name,
	}

	code, status, err := c.get(ctx, "/health")
	// Apps may still have /healthcheck endpoint instead of a /health one.
	if code == http.StatusNotFound {
		code, status, err = c.get(ctx, "/healthcheck")
	}
	if err != nil {
		errorMessage = err.Error()
		log.Event(ctx, "failed to request service health", log.Error(err), logData)
	}

	check := getCheck(ctx, c.Name, status, errorMessage, code)

	return check, nil
}

func (c *Client) get(ctx context.Context, path string) (int, string, error) {
	var check *health.HealthCheck
	clientlog.Do(ctx, "checking health", c.Name, path)

	req, err := http.NewRequest("GET", c.URL+path, nil)
	if err != nil {
		return 0, health.StatusCritical, err
	}

	resp, err := c.Client.Do(ctx, req)
	if err != nil {
		return 0, health.StatusCritical, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		if resp.Body != nil {
			io.Copy(ioutil.Discard, resp.Body)
		}
		return resp.StatusCode, health.StatusCritical, ErrInvalidAPIResponse{http.StatusOK, resp.StatusCode, req.URL.Path}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, health.StatusCritical, err
	}

	if err = json.Unmarshal(b, &check); err != nil {
		return resp.StatusCode, check.Status, err
	}

	return resp.StatusCode, check.Status, nil
}
