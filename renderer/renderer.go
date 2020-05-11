package renderer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/clientlog"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/log"
)

const service = "renderer"

// ErrInvalidRendererResponse is returned when the renderer service does not respond
// with a status 200
type ErrInvalidRendererResponse struct {
	responseCode int
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidRendererResponse) Error() string {
	return fmt.Sprintf("invalid response from renderer service - status %d", e.responseCode)
}

// Code returns the status code received from renderer if an error is returned
func (e ErrInvalidRendererResponse) Code() int {
	return e.responseCode
}

// Renderer represents a renderer client to interact with the dp-frontend-renderer
type Renderer struct {
	cli dphttp.Clienter
	url string
}

// New creates an instance of renderer with a default client
func New(url string) *Renderer {
	hcClient := healthcheck.NewClient(service, url)

	return &Renderer{
		cli: hcClient.Client,
		url: url,
	}
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Event(ctx, "error closing http response body", log.ERROR, log.Error(err))
	}
}

// Checker calls dataset api health endpoint and returns a check object to the caller.
func (r *Renderer) Checker(ctx context.Context, check *health.CheckState) error {
	hcClient := healthcheck.Client{
		Client: r.cli,
		URL:    r.url,
		Name:   service,
	}

	return hcClient.Checker(ctx, check)
}

// Do sends a request to the renderer service to render a given template
func (r *Renderer) Do(path string, b []byte) ([]byte, error) {
	// Renderer required JSON to be sent so if byte array is empty, set it to be
	// empty json
	if b == nil {
		b = []byte(`{}`)
	}

	uri := r.url + "/" + path
	ctx := context.Background()

	clientlog.Do(ctx, fmt.Sprintf("rendering template: %s", path), service, uri, log.Data{
		"method": "POST",
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.cli.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidRendererResponse{resp.StatusCode}
	}

	return ioutil.ReadAll(resp.Body)
}
