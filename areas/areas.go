package areas

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "areas-api"

// ErrInvalidAreaAPIResponse is returned when the area api does not respond
// with a valid status
type ErrInvalidAreaAPIResponse struct {
	actualCode int
	uri        string
	body       string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidAreaAPIResponse) Error() string {
	return fmt.Sprintf("invalid response: %d from dataset api: %s, body: %s",
		e.actualCode,
		e.uri,
		e.body,
	)
}

// Code returns the status code received from Area api if an error is returned
func (e ErrInvalidAreaAPIResponse) Code() int {
	return e.actualCode
}

// Client is a areas api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// New creates a new instance of Client with a given areas api url
func New(areasAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, areasAPIURL),
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls areas api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// GetArea returns area information for a given area ID
func (c *Client) GetArea(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, areaID, acceptLang string) (areaDetails AreaDetails, err error) {
	uri := fmt.Sprintf("%s/v1/areas/%s", c.hcCli.URL, areaID)
	clientlog.Do(ctx, "retrieving area", service, uri)
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil, "", acceptLang)
	if err != nil {
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = NewAreaAPIResponse(resp, uri)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var body map[string]interface{}
	if err = json.Unmarshal(b, &body); err != nil {
		return
	}

	// TODO: Authentication will sort this problem out for us. Currently
	// the shape of the response body is different if you are authenticated
	// so return the "next" item only
	if next, ok := body["next"]; ok && (serviceAuthToken != "" || userAuthToken != "") {
		b, err = json.Marshal(next)
		if err != nil {
			return
		}
	}

	err = json.Unmarshal(b, &areaDetails)
	return
}

// GetRelations gets the child areas
func (c *Client) GetRelations(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, areaID, acceptLang string) (relations []Relation, err error) {
	uri := fmt.Sprintf("%s/v1/areas/%s/relations?relationship=child", c.hcCli.URL, areaID)
	clientlog.Do(ctx, "retrieving child areas relations", service, uri)
	// Do request
	res, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, collectionID, uri, nil, "", acceptLang)
	if err != nil {
		return
	}

	defer closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusOK {
		err = NewAreaAPIResponse(res, uri)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &relations)
	return
}

// NewAreaAPIResponse creates an error response, optionally adding body to e when status is 404
func NewAreaAPIResponse(resp *http.Response, uri string) (e *ErrInvalidAreaAPIResponse) {
	e = &ErrInvalidAreaAPIResponse{
		actualCode: resp.StatusCode,
		uri:        uri,
	}
	if resp.StatusCode == http.StatusNotFound {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			e.body = "Client failed to read area body"
			return
		}
		defer closeResponseBody(nil, resp)

		e.body = string(b)
	}
	return
}

func addCollectionIDHeader(r *http.Request, collectionID string) {
	if len(collectionID) > 0 {
		r.Header.Add(dprequest.CollectionIDHeaderKey, collectionID)
	}
}

// doGetWithAuthHeaders executes a GET request by using clienter.Do for the provided URI and payload body.
// It sets the user and service authentication and collectionID as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
// If url.Values are provided, they will be added as query parameters in the URL.
// NOTE: Only one of the tokens 'userAuthToken' or 'serviceAuthToken' needs to have a value.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, uri string, values url.Values, ifMatch, acceptLang string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if values != nil {
		req.URL.RawQuery = values.Encode()
	}

	headers.SetIfMatch(req, ifMatch)
	headers.SetAcceptedLang(req, acceptLang)
	addCollectionIDHeader(req, collectionID)
	dprequest.AddFlorenceHeader(req, userAuthToken)
	dprequest.AddServiceTokenHeader(req, serviceAuthToken)
	return c.hcCli.Client.Do(ctx, req)
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
