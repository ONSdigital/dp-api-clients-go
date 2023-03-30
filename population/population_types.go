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

type PopulationType struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type GetPopulationTypeInput struct {
	PopulationType string
	AuthTokens
}

type GetPopulationTypesInput struct {
	PaginationParams
	DefaultDatasets bool
	AuthTokens
}

type GetPopulationTypeResponse struct {
	PopulationType PopulationType `json:"population_type"`
}

type GetPopulationTypesResponse struct {
	Items []PopulationType `json:"items"`
}

func (c *Client) GetPopulationTypes(ctx context.Context, input GetPopulationTypesInput) (GetPopulationTypesResponse, error) {
	logData := log.Data{
		"method": http.MethodGet,
	}

	urlPath := "population-types"
	if input.Limit > 0 {
		urlPath = fmt.Sprintf("population-types?limit=%d&offset=%d", input.Limit, input.Offset)
	}

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

func (c *Client) GetPopulationType(ctx context.Context, input GetPopulationTypeInput) (GetPopulationTypeResponse, error) {
	logData := log.Data{
		"method": http.MethodGet,
	}

	urlPath := "population-types/" + input.PopulationType

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, nil)
	if err != nil {
		return GetPopulationTypeResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting population type", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetPopulationTypeResponse{}, dperrors.New(
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
		return GetPopulationTypeResponse{}, err
	}

	var resp GetPopulationTypeResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetPopulationTypeResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize population types response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}
