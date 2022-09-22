package population

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

// Area is an area model with ID and Label
type Area struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AreaType string `json:"area_type"`
}

type GetAreaInput struct {
	AuthTokens
	PopulationType string
	AreaType       string
	Area           string
}

type GetAreasInput struct {
	AuthTokens
	PopulationType string
	AreaTypeID     string
	Text           string
}

type GetParentAreaCountInput struct {
	AuthTokens
	PopulationType   string
	AreaTypeID       string
	ParentAreaTypeID string
	Areas            []string
}

// GetAreasResponse is the response object for GET /areas
type GetAreasResponse struct {
	PaginationResponse
	Areas []Area `json:"areas"`
}

// GetAreasResponse is the response object for GET /areas
type GetAreaResponse struct {
	Area Area `json:"area"`
}

func (c *Client) GetArea(ctx context.Context, input GetAreaInput) (GetAreaResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"population_type": input.PopulationType,
		"area_type":       input.AreaType,
		"area":            input.Area,
	}
	urlPath := fmt.Sprintf("population-types/%s/area-types/%s/areas/%s", input.PopulationType, input.AreaType, input.Area)

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, nil)
	if err != nil {
		return GetAreaResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting area", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreaResponse{}, dperrors.New(
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
		return GetAreaResponse{}, err
	}

	var resp GetAreaResponse

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return GetAreaResponse{}, err
	}

	logData["resp"] = string(b)

	if err := json.Unmarshal(b, &resp); err != nil {
		return GetAreaResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize area response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}

func (c *Client) GetAreas(ctx context.Context, input GetAreasInput) (GetAreasResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"population_type": input.PopulationType,
		"area_type_id":    input.AreaTypeID,
		"text":            input.Text,
	}

	urlPath := fmt.Sprintf("population-types/%s/area-types/%s/areas", input.PopulationType, input.AreaTypeID)
	var urlValues map[string][]string
	if input.Text != "" {
		urlValues = url.Values{"q": []string{input.Text}}
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetAreasResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting areas", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetAreasResponse{}, dperrors.New(
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
		return GetAreasResponse{}, err
	}

	var areas GetAreasResponse

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return GetAreasResponse{}, err
	}

	if err := json.Unmarshal(b, &areas); err != nil {
		return GetAreasResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize areas response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return areas, nil
}

func (c *Client) GetParentAreaCount(ctx context.Context, input GetParentAreaCountInput) (int, error) {
	logData := log.Data{
		"method":              http.MethodGet,
		"dataset_id":          input.PopulationType,
		"area_type_id":        input.AreaTypeID,
		"parent_area_type_id": input.ParentAreaTypeID,
		"areas":               input.Areas,
	}

	urlPath := fmt.Sprintf("population-types/%s/area-types/%s/parents/%s/areas-count",
		input.PopulationType,
		input.AreaTypeID,
		input.ParentAreaTypeID,
	)

	urlValues := map[string][]string{"areas": {strings.Join(input.Areas, ",")}}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return 0, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting area-types parents", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return 0, dperrors.New(
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
		return 0, err
	}

	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, dperrors.New(
			errors.Wrap(err, "unable to deserialize parent areas count response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	if err != nil {
		return 0, dperrors.New(
			errors.Wrap(err, "unable to convert parent areas count API response"),
			http.StatusInternalServerError,
			logData,
		)
	}
	return count, nil
}
