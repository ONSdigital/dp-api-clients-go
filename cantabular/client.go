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

// Cantabular service names
const (
	Service         = "cantabular"
	ServiceAPIExt   = "cantabularAPIExt"
	ServiceMetadata = "cantabularMetadataService"
	SoftwareVersion = "v10"
)

var (
	tableErrors = map[string]string{
		"withinMaxCells": "resulting dataset too large",
	}
)

// Client is the client for interacting with the Cantabular API
type Client struct {
	ua         httpClient
	gqlClient  GraphQLClient
	host       string
	extApiHost string
	version    string
}

// NewClient returns a new Client
func NewClient(cfg Config, ua httpClient, g GraphQLClient) *Client {
	c := &Client{
		ua:         ua,
		gqlClient:  g,
		host:       cfg.Host,
		extApiHost: cfg.ExtApiHost,
		version:    SoftwareVersion,
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
				"method": "GET",
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

// Checker contacts the /vXX/datasets endpoint and updates the healthcheck state accordingly.
func (c *Client) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	reqURL := fmt.Sprintf("%s/%s/datasets", c.host, c.version)
	return c.checkHealth(ctx, state, Service, reqURL)
}

// CheckerAPIExt contacts the /graphql endpoint with an empty query and updates the healthcheck state accordingly.
func (c *Client) CheckerAPIExt(ctx context.Context, state *healthcheck.CheckState) error {
	reqURL := fmt.Sprintf("%s/graphql?query={datasets{name}}", c.extApiHost)
	return c.checkHealth(ctx, state, ServiceAPIExt, reqURL)
}

// CheckerMetadataService contacts the /graphql endpoint and updates the healthcheck state accordingly.
func (c *Client) CheckerMetadataService(ctx context.Context, state *healthcheck.CheckState) error {
	// FIXME: We should not be using ext api host but that is the host used to create the graphql client
	// despite it actually containing the dp-cantabular-metadata-service url as a value
	reqURL := fmt.Sprintf("%s/graphql", c.extApiHost)
	return c.checkHealth(ctx, state, ServiceMetadata, reqURL)
}

func (c *Client) checkHealth(ctx context.Context, state *healthcheck.CheckState, service, reqURL string) error {
	logData := log.Data{
		"service": service,
	}
	code := 0

	res, err := c.httpGet(ctx, reqURL)
	defer closeResponseBody(ctx, res)

	if err != nil {
		log.Error(ctx, "failed to request service health", err, logData)
	} else {
		code = res.StatusCode
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

// StatusCode provides a callback function whereby users can check a returned
// error for an embedded HTTP status code
func (c *Client) StatusCode(err error) int {
	var cerr coder
	if errors.As(err, &cerr) {
		return cerr.Code()
	}

	return 0
}

func (c *Client) parseTableError(err string) string {
	if tErr, ok := tableErrors[err]; ok {
		return tErr
	}

	return err
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
