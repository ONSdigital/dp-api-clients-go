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
	CheckObj *health.Check
	Client   rchttp.Clienter
	Name     string
	URL      string
}

// NewClient creates a new instance of Client with a given app url
func NewClient(name, url string) *Client {
	c := &Client{
		Client: rchttp.NewClient(),
		Name:   name,
		URL:    url,
		CheckObj: &health.Check{
			Name: name,
		},
	}

	// healthcheck client should not retry when calling a healthcheck endpoint,
	// append to current paths as to not change the client setup by service
	paths := c.Client.GetPathsWithNoRetries()
	paths = append(paths, "/health", "/healthcheck")
	c.Client.SetPathsWithNoRetries(paths)

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
func (c *Client) Checker(ctx context.Context) (*health.Check, error) {
	logData := log.Data{
		"service": c.Name,
	}

	code, err := c.get(ctx, "/health")
	// Apps may still have /healthcheck endpoint
	// instead of a /health one
	if code == http.StatusNotFound {
		code, err = c.get(ctx, "/healthcheck")
	}
	if err != nil {
		log.Event(ctx, "failed to request api health", log.Error(err), logData)
	}

	currentTime := time.Now().UTC()
	c.CheckObj.StatusCode = code
	c.CheckObj.LastChecked = &currentTime

	switch code {
	case 200:
		c.CheckObj.Message = statusDescription[health.StatusOK]
		c.CheckObj.Status = health.StatusOK
		c.CheckObj.LastSuccess = &currentTime
	case 429:
		c.CheckObj.Message = statusDescription[health.StatusWarning]
		c.CheckObj.Status = health.StatusWarning
		c.CheckObj.LastFailure = &currentTime
	default:
		c.CheckObj.Message = statusDescription[health.StatusCritical]
		c.CheckObj.Status = health.StatusCritical
		c.CheckObj.LastFailure = &currentTime
	}

	return c.CheckObj, nil
}

func (c *Client) get(ctx context.Context, path string) (int, error) {
	req, err := http.NewRequest("GET", c.URL+path, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.Client.Do(ctx, req)
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
