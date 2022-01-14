package cantabular

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

// Codebook represents a 'codebook' object returned from Cantabular Server
type Codebook []Variable

// GetCodebook gets a Codebook from cantabular.
func (c *Client) GetCodebook(ctx context.Context, req GetCodebookRequest) (*GetCodebookResponse, error) {
	if len(c.host) == 0 {
		return nil, dperrors.New(
			errors.New("cantabular server host not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	var vars string
	for _, v := range req.Variables {
		vars += "&v=" + v
	}

	url := fmt.Sprintf("%s/v9/codebook/%s?cats=%v%s", c.host, req.DatasetName, req.Categories, vars)

	res, err := c.httpGet(ctx, url)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to get response from Cantabular API: %s", err),
			http.StatusInternalServerError,
			log.Data{
				"url": url,
			},
		)
	}
	defer closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusOK {
		return nil, c.errorResponse(url, res)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			res.StatusCode,
			log.Data{
				"url": url,
			},
		)
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	var resp GetCodebookResponse

	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{
				"url":           url,
				"response_body": string(b),
			},
		)
	}

	return &resp, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp != nil && resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
