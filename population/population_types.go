package population

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

type PopulationType struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type GetPopulationTypesInput struct {
	DefaultDatasets bool
	AuthTokens
}

type GetPopulationTypesResponse struct {
	Items []PopulationType `json:"items"`
}

func (c *Client) GetPopulationTypes(ctx context.Context, input GetPopulationTypesInput) (GetPopulationTypesResponse, error) {
	logData := log.Data{
		"method": http.MethodGet,
	}

	urlPath := "population-types"
	urlValues := url.Values{}
	if input.DefaultDatasets {
		urlValues.Add("require-default-dataset", "true")
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetPopulationTypesResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting population types", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetPopulationTypesResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Population types API"),
			http.StatusInternalServerError,
			logData,
		)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}()

	if err := checkGetResponse(res); err != nil {
		return GetPopulationTypesResponse{}, err
	}

	var resp GetPopulationTypesResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetPopulationTypesResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize population types response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}
