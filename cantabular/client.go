package cantabular

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/shurcooL/graphql"
)

// Service is the cantabular service name
const Service = "cantabular"
const ServiceApiExt = "cantabularApiExt"

// Client is the client for interacting with the Cantabular API
type Client struct {
	ua         httpClient
	gqlClient  GraphQLClient
	host       string
	extApiHost string
}

// NewClient returns a new Client
func NewClient(cfg Config, ua httpClient, g GraphQLClient) *Client {
	c := &Client{
		ua:         ua,
		gqlClient:  g,
		host:       cfg.Host,
		extApiHost: cfg.ExtApiHost,
	}

	if len(cfg.ExtApiHost) > 0 && c.gqlClient == nil {
		c.gqlClient = graphql.NewClient(
			fmt.Sprintf("%s/graphql", cfg.ExtApiHost),
			&http.Client{
				Timeout: cfg.GraphQLTimeout,
			},
		)
	}

	return c
}

// httpGet makes a get request to the given url and returns the response
func (c *Client) httpGet(ctx context.Context, path string) (*http.Response, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to parse url: %s", err),
			http.StatusBadRequest,
			log.Data{
				"url": path,
			},
		)
	}

	path = URL.String()

	resp, err := c.ua.Get(ctx, path)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make request: %w", err),
			http.StatusInternalServerError,
			log.Data{
				"url":    path,
				"method": "get",
			},
		)
	}

	return resp, nil
}

// httpPost makes a post request to the given url and returns the response
func (c *Client) httpPost(ctx context.Context, path string, contentType string, body io.Reader) (*http.Response, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to parse url: %s", err),
			http.StatusBadRequest,
			log.Data{
				"url": path,
			},
		)
	}

	path = URL.String()

	resp, err := c.ua.Post(ctx, path, contentType, body)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make request: %w", err),
			http.StatusInternalServerError,
			log.Data{
				"url":    path,
				"method": "post",
			},
		)
	}

	return resp, nil
}

// Checker contacts the /v9/datasets endpoint and updates the healthcheck state accordingly.
func (c *Client) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	reqURL := fmt.Sprintf("%s/v9/datasets", c.host)
	return c.checkHealth(ctx, state, Service, reqURL)
}

// CheckerApiExt contacts the /graphql endpoint with an empty query and updates the healthcheck state accordingly.
func (c *Client) CheckerApiExt(ctx context.Context, state *healthcheck.CheckState) error {
	reqURL := fmt.Sprintf("%s/graphql?query={}", c.extApiHost)
	return c.checkHealth(ctx, state, ServiceApiExt, reqURL)
}

func (c *Client) checkHealth(ctx context.Context, state *healthcheck.CheckState, service, reqURL string) error {
	logData := log.Data{
		"service": service,
	}
	code := 0

	res, err := c.httpGet(ctx, reqURL)
	if err != nil {
		log.Error(ctx, "failed to request service health", err, logData)
	} else {
		code = res.StatusCode
		defer closeResponseBody(ctx, res)
	}

	switch code {
	case 0: // When there is a problem with the client return error in message
		return state.Update(healthcheck.StatusCritical, err.Error(), 0)
	case 200:
		message := service + health.StatusMessage[healthcheck.StatusOK]
		return state.Update(healthcheck.StatusOK, message, code)
	default:
		message := service + health.StatusMessage[healthcheck.StatusCritical]
		return state.Update(healthcheck.StatusCritical, message, code)
	}
}

// errorResponse handles dealing with an error response from Cantabular
func (c *Client) errorResponse(url string, res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read error response body: %s", err),
			res.StatusCode,
			log.Data{
				"url": url,
			},
		)
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var resp ErrorResponse

	if err := json.Unmarshal(b, &resp); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal error response body: %s", err),
			res.StatusCode,
			log.Data{
				"url":           url,
				"response_body": string(b),
			},
		)
	}

	return dperrors.New(
		errors.New(resp.Message),
		res.StatusCode,
		log.Data{
			"url": url,
		},
	)
}
