package health

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/log.go/log"
)

var (
	statusDescription = map[string]string{
		health.StatusOK:       "Everything is ok",
		health.StatusWarning:  "Things are degraded, but at least partially functioning",
		health.StatusCritical: "The checked functionality is unavailable or non-functioning",
	}
)

// ErrInvalidAppResponse is returned when an app does not respond
// with a valid status
type ErrInvalidAppResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Client represents an app client
type Client struct {
	checkObj *health.Check
	client   rchttp.Clienter
	name     string
	url      string
}

// NewClient creates a new instance of Client with a given app url
func NewClient(name, url string, maxRetries int) *Client {
	c := &Client{
		client: rchttp.NewClient(),
		name:   name,
		url:    url,
		checkObj: &health.Check{
			Name: name,
		},
	}

	// Overwrite the default number of max retries on the new client
	c.client.SetMaxRetries(maxRetries)

	return c
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAppResponse) Error() string {
	return fmt.Sprintf("invalid response from downstream service - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Checker calls an app health endpoint and returns a check object to the caller
func (c *Client) Checker(ctx *context.Context) (*health.Check, error) {
	logData := log.Data{
		"service": c.name,
	}

	code, err := c.get(*ctx, "/health")
	// Apps may still have /healthcheck endpoint
	// instead of a /health one
	if code == http.StatusNotFound {
		code, err = c.get(*ctx, "/healthcheck")
	}
	if err != nil {
		log.Event(*ctx, "failed to request api health", log.Error(err), logData)
	}

	currentTime := time.Now().UTC()
	c.checkObj.StatusCode = code
	c.checkObj.LastChecked = &currentTime

	switch code {
	case 200:
		c.checkObj.Message = statusDescription[health.StatusOK]
		c.checkObj.Status = health.StatusOK
		c.checkObj.LastSuccess = &currentTime
	case 429:
		c.checkObj.Message = statusDescription[health.StatusWarning]
		c.checkObj.Status = health.StatusWarning
		c.checkObj.LastFailure = &currentTime
	default:
		c.checkObj.Message = statusDescription[health.StatusCritical]
		c.checkObj.Status = health.StatusCritical
		c.checkObj.LastFailure = &currentTime
	}

	return c.checkObj, nil
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
		return resp.StatusCode, ErrInvalidAppResponse{http.StatusOK, resp.StatusCode, req.URL.Path}
	}

	return resp.StatusCode, nil
}
