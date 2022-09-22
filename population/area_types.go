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

// AreaType is an area type model with ID and Label
type AreaType struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	TotalCount int    `json:"total_count"`
}

type GetAreaTypeParentsInput struct {
	AuthTokens
	PopulationType string
	AreaTypeID     string
}

// GetAreaTypesResponse is the response object for GET /area-types
type GetAreaTypesResponse struct {
	PaginationResponse
	AreaTypes []AreaType `json:"area_types"`
}

// GetAreaTypeParentsResponse is the response object for GET /areas
type GetAreaTypeParentsResponse struct {
	PaginationResponse
	AreaTypes []AreaType `json:"area_types"`
}

// GetPopulationAreaTypes retrieves the Cantabular area-types associated with a dataset
func (c *Client) GetAreaTypes(ctx context.Context, userAuthToken, serviceAuthToken, datasetID string) (GetAreaTypesResponse, error) {
	logData := log.Data{
		"method":     http.MethodGet,
		"dataset_id": datasetID,
	}

	urlPath := fmt.Sprintf("population-types/%s/area-types", datasetID)
	urlValues := url.Values{}

	req, err := c.createGetRequest(ctx, userAuthToken, serviceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting area types", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Population Type API"),
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
		return GetAreaTypesResponse{}, err
	}

	var areaTypes GetAreaTypesResponse
	if err := json.NewDecoder(resp.Body).Decode(&areaTypes); err != nil {
		return GetAreaTypesResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize area types response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return areaTypes, nil
}

func (c *Client) GetAreaTypeParents(ctx context.Context, input GetAreaTypeParentsInput) (GetAreaTypeParentsResponse, error) {
	logData := log.Data{
		"method":       http.MethodGet,
		"dataset_id":   input.PopulationType,
		"area_type_id": input.AreaTypeID,
	}

	urlPath := fmt.Sprintf("population-types/%s/area-types/%s/parents", input.PopulationType, input.AreaTypeID)
	var urlValues map[string][]string

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetAreaTypeParentsResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting area-types parents", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreaTypeParentsResponse{}, dperrors.New(
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
		return GetAreaTypeParentsResponse{}, err
	}

	var atp GetAreaTypeParentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&atp); err != nil {
		return GetAreaTypeParentsResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize areas response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return atp, nil
}
