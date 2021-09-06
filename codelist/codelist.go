package codelist

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/log.go/v2/log"
)

const service = "code-list-api"

var _ error = ErrInvalidCodelistAPIResponse{}

// Client is a codelist api client which can be used to make requests to the server
type Client struct {
	hcCli *healthcheck.Client
}

// ErrInvalidCodelistAPIResponse is returned when the codelist api does not respond
// with a valid status
type ErrInvalidCodelistAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidCodelistAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from codelist api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from code list api if an error is returned
func (e ErrInvalidCodelistAPIResponse) Code() int {
	return e.actualCode
}

// New creates a new instance of Client with a given filter api url
func New(codelistAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, codelistAPIURL),
	}
}

// NewWithHealthClient creates a new instance of CodelistAPI Client,
// reusing the URL and Clienter from the provided healthcheck client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// URL returns the URL used by this client
func (c *Client) URL() string {
	return c.hcCli.URL
}

// HealthClient returns the underlying Healthcheck Client for this codelistAPI client
func (c *Client) HealthClient() *healthcheck.Client {
	return c.hcCli
}

// Checker calls codelist api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// GetValues returns dimension values from the codelist api
func (c *Client) GetValues(ctx context.Context, userAuthToken string, serviceAuthToken string, id string) (DimensionValues, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.hcCli.URL, id)
	clientlog.Do(ctx, "retrieving codes from codelist", service, uri)

	var vals DimensionValues
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return vals, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return vals, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return vals, err
	}

	err = json.Unmarshal(b, &vals)
	return vals, err
}

// GetIDNameMap returns dimension values in the form of an id name map
func (c *Client) GetIDNameMap(ctx context.Context, userAuthToken string, serviceAuthToken string, id string) (map[string]string, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.hcCli.URL, id)
	clientlog.Do(ctx, "retrieving codes from codelist for id name map", service, uri)

	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vals DimensionValues
	if err = json.Unmarshal(body, &vals); err != nil {
		return nil, err
	}

	idNames := make(map[string]string)
	for _, val := range vals.Items {
		idNames[val.Code] = val.Label
	}

	return idNames, nil
}

// GetGeographyCodeLists returns the geography codelists
func (c *Client) GetGeographyCodeLists(ctx context.Context, userAuthToken string, serviceAuthToken string) (CodeListResults, error) {
	uri := fmt.Sprintf("%s/code-lists?type=geography", c.hcCli.URL)
	clientlog.Do(ctx, "retrieving geography codelists", service, uri)

	var results CodeListResults
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return results, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return results, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return results, err
	}

	err = json.Unmarshal(b, &results)
	if err != nil {
		return results, err
	}
	return results, nil
}

// GetCodeListEditions returns the editions for a codelist
func (c *Client) GetCodeListEditions(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string) (EditionsListResults, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/editions", c.hcCli.URL, codeListID)
	clientlog.Do(ctx, "retrieving codelist editions", service, uri)

	var editionsList EditionsListResults
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return editionsList, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != 200 {
		return editionsList, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return editionsList, err
	}

	err = json.Unmarshal(b, &editionsList)
	if err != nil {
		return editionsList, err
	}

	return editionsList, nil
}

// GetCodes returns the codes for a specific edition of a code list
func (c *Client) GetCodes(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string) (CodesResults, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes", c.hcCli.URL, codeListID, edition)
	clientlog.Do(ctx, "retrieving codes from an edition of a code list", service, uri)

	var codes CodesResults
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return codes, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return codes, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return codes, err
	}

	err = json.Unmarshal(b, &codes)
	if err != nil {
		return codes, err
	}

	return codes, nil
}

// GetCodeByID returns information about a code
func (c *Client) GetCodeByID(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (CodeResult, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s", c.hcCli.URL, codeListID, edition, codeID)
	clientlog.Do(ctx, "retrieving code from an edition of a code list", service, uri)

	var code CodeResult
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return code, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return code, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return code, err
	}

	err = json.Unmarshal(b, &code)
	if err != nil {
		return code, err
	}

	return code, nil
}

// GetDatasetsByCode returns datasets containing the codelist codeID.
func (c *Client) GetDatasetsByCode(ctx context.Context, userAuthToken string, serviceAuthToken string, codeListID string, edition string, codeID string) (DatasetsResult, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s/datasets", c.hcCli.URL, codeListID, edition, codeID)
	clientlog.Do(ctx, "retrieving datasets containing a code from an edition of a code list", service, uri)

	var datasets DatasetsResult
	resp, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return datasets, err
	}

	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return datasets, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return datasets, err
	}

	err = json.Unmarshal(b, &datasets)
	if err != nil {
		return datasets, err
	}
	return datasets, nil
}

// doGetWithAuthHeaders executes clienter.Do setting the service authentication token as a request header. Returns the http.Response and any error.
// It is the callers responsibility to ensure response.Body is closed on completion.
func (c *Client) doGetWithAuthHeaders(ctx context.Context, userAuthToken string, serviceAuthToken string, uri string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	if err := setAuthenticationHeaders(req, userAuthToken, serviceAuthToken); err != nil {
		return nil, err
	}

	return c.hcCli.Client.Do(ctx, req)
}

func setAuthenticationHeaders(req *http.Request, userAuthToken, serviceAuthToken string) error {
	err := headers.SetAuthToken(req, userAuthToken)
	if err != nil && err != headers.ErrValueEmpty {
		return err
	}

	err = headers.SetServiceAuthToken(req, serviceAuthToken)
	if err != nil && err != headers.ErrValueEmpty {
		return err
	}

	return nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
