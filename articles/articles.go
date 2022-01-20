package articles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/log.go/v2/log"
)

// GetLegacyBulletin returns a legacy bulletin
func (c *Client) GetLegacyBulletin(ctx context.Context, userAccessToken, collectionID, lang, uri string) (*Bulletin, error) {
	url := fmt.Sprintf("%s/articles/legacy?url=%s&lang=%s", c.hcCli.URL, uri, lang)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to create request to Articles API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}

	headers.SetCollectionID(req, collectionID)
	headers.SetAuthToken(req, userAccessToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to get response from Articles API: %s", err),
			http.StatusInternalServerError,
			nil,
		)
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, c.errorResponse(resp)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to read response body from Articles API: %s", err),
			resp.StatusCode,
			nil,
		)
	}

	var bulletin Bulletin
	if err = json.Unmarshal(b, &bulletin); err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			log.Data{"response_body": string(b)},
		)
	}

	return &bulletin, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
