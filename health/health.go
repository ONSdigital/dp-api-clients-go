package health

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

// ErrInvalidAPIResponse is returned when an api does not respond
// with a valid status
type ErrInvalidAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Client represents an api client
type Client struct {
	client rchttp.Clienter
	name   string
	url    string
}

// NewClient creates a new instance of Client with a given api url
func NewClient(name, url string, maxRetries int) *Client {
	c := &Client{
		client: rchttp.NewClient(),
		name:   name,
		url:    url,
	}

	// Overwrite the default number of max retries on the new client
	c.client.SetMaxRetries(maxRetries)

	return c
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from downstream api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Checker calls an api health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx *context.Context) (*health.Check, error) {
	logData := log.Data{
		"api": c.name,
	}

	statusCode, err := c.get(*ctx, "/health")
	// Apps may still have /healthcheck endpoint
	// instead of a /health one
	if statusCode == http.StatusNotFound {
		statusCode, err = c.get(*ctx, "/healthcheck")
	}
	if err != nil {
		log.Event(*ctx, "failed to request api health", log.Error(err), logData)
	}

	check := getCheck(ctx, c.name, statusCode)

	return check, nil
}

func (c *Client) get(ctx context.Context, path string) (int, error) {
	req, err := http.NewRequest("GET", c.url+path, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || (resp.StatusCode > 399 && resp.StatusCode != 429) {
		io.Copy(ioutil.Discard, resp.Body)
		return resp.StatusCode, ErrInvalidAPIResponse{http.StatusOK, resp.StatusCode, req.URL.Path}
	}

	return resp.StatusCode, nil
}
