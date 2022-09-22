package population

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

// Dimension is an area-type model with ID and Label
type Dimension struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	TotalCount int    `json:"total_count"`
}

type GetDimensionsInput struct {
	AuthTokens
	PaginationParams
	PopulationType string
	SearchString   string
}

// GetDimensionsResponse is the response object for GetDimensions
type GetDimensionsResponse struct {
	PaginationResponse
	Dimensions []Dimension `json:"items"`
}

func (c *Client) GetDimensions(ctx context.Context, input GetDimensionsInput) (GetDimensionsResponse, error) {
	logData := log.Data{
		"method":         http.MethodGet,
		"limit":          input.Limit,
		"offset":         input.Offset,
		"populationType": input.PopulationType,
		"search string":  input.SearchString,
	}

	urlPath := fmt.Sprintf("/population-types/%s/dimensions", input.PopulationType)
	var urlValues map[string][]string
	if input.SearchString != "" {
		urlValues = url.Values{"q": []string{input.SearchString}}
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetDimensionsResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting dimensions", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetDimensionsResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Population types API"),
			http.StatusInternalServerError,
			logData,
		)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}()

	if err := checkGetResponse(resp); err != nil {
		return GetDimensionsResponse{}, err
	}

	var dimensions GetDimensionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dimensions); err != nil {
		return GetDimensionsResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize areas response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return dimensions, nil
}
