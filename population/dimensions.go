package population

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

// Dimension is an area-type model with ID and Label
type Dimension struct {
	ID                   string     `json:"id"`
	Label                string     `json:"label"`
	Description          string     `json:"description"`
	Categories           []Category `json:"categories"`
	TotalCount           int        `json:"total_count"`
	QualityStatementText string     `json:"quality_statement_text"`
}

type DimensionCategory struct {
	Id         string                  `json:"id"`
	Label      string                  `json:"label"`
	Categories []DimensionCategoryItem `json:"categories"`
}

type DimensionCategoryItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Category struct {
	ID                   string `json:"id"`
	Label                string `json:"label"`
	QualityStatementText string `json:"quality_statement_text"`
}

type GetDimensionCategoryInput struct {
	AuthTokens
	PaginationParams
	PopulationType string
	Dimensions     []string
}

type GetDimensionsInput struct {
	AuthTokens
	PaginationParams
	PopulationType string
	SearchString   string
}

type GetDimensionsDescriptionInput struct {
	AuthTokens
	PopulationType string
	DimensionIDs   []string
}

type GetCategorisationsInput struct {
	AuthTokens
	PaginationParams
	PopulationType string
	Dimension      string
}

type GetDimensionCategoriesResponse struct {
	PaginationResponse
	Categories []DimensionCategory `json:"items"`
}

// GetDimensionsResponse is the response object for GetDimensions
type GetDimensionsResponse struct {
	PaginationResponse
	Dimensions []Dimension `json:"items"`
}

type GetCategorisationsResponse struct {
	PaginationResponse
	Items []Dimension `json:"items"`
}

type GetBaseVariableInput struct {
	AuthTokens
	PopulationType string
	Variable       string
}

type GetBaseVariableResponse struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func (c *Client) GetDimensionCategories(ctx context.Context, input GetDimensionCategoryInput) (GetDimensionCategoriesResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"limit":           input.Limit,
		"offset":          input.Offset,
		"population_type": input.PopulationType,
		"dimensions":      input.Dimensions,
	}

	urlPath := fmt.Sprintf("population-types/%s/dimension-categories", input.PopulationType)
	dimensionsString := strings.Join(input.Dimensions[:], ",")
	urlValues := url.Values{
		"dims":   []string{dimensionsString},
		"limit":  []string{strconv.Itoa(input.Limit)},
		"offset": []string{strconv.Itoa(input.Offset)},
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetDimensionCategoriesResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting dimension categories", service, req.URL.String(), logData)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetDimensionCategoriesResponse{}, dperrors.New(
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
		return GetDimensionCategoriesResponse{}, err
	}

	var dimensionCategories GetDimensionCategoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&dimensionCategories); err != nil {
		return GetDimensionCategoriesResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize areas response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return dimensionCategories, nil

}
func (c *Client) GetDimensions(ctx context.Context, input GetDimensionsInput) (GetDimensionsResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"limit":           input.Limit,
		"offset":          input.Offset,
		"population_type": input.PopulationType,
		"search_string":   input.SearchString,
	}

	urlPath := fmt.Sprintf("population-types/%s/dimensions", input.PopulationType)
	urlValues := url.Values{
		"limit":  []string{strconv.Itoa(input.Limit)},
		"offset": []string{strconv.Itoa(input.Offset)},
	}
	if input.SearchString != "" {
		urlValues["q"] = []string{input.SearchString}
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

func (c *Client) GetDimensionsDescription(ctx context.Context, input GetDimensionsDescriptionInput) (GetDimensionsResponse, error) {
	logData := log.Data{
		"method":       http.MethodGet,
		"dimensionIDs": input.DimensionIDs,
	}

	urlPath := fmt.Sprintf("population-types/%s/dimensions-description", input.PopulationType)

	var urlValues url.Values
	if input.DimensionIDs != nil {
		urlValues = make(url.Values)
		urlValues["q"] = input.DimensionIDs
	}

	req, err := c.createGetDimensionsDescriptionRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
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

func (c *Client) GetCategorisations(ctx context.Context, input GetCategorisationsInput) (GetCategorisationsResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"limit":           input.Limit,
		"offset":          input.Offset,
		"population_type": input.PopulationType,
		"dimension":       input.Dimension,
	}

	urlPath := fmt.Sprintf("population-types/%s/dimensions/%s/categorisations", input.PopulationType, input.Dimension)
	urlValues := url.Values{
		"limit":  []string{strconv.Itoa(input.Limit)},
		"offset": []string{strconv.Itoa(input.Offset)},
	}

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, urlValues)
	if err != nil {
		return GetCategorisationsResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting dimension categorisations", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetCategorisationsResponse{}, dperrors.New(
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
		return GetCategorisationsResponse{}, err
	}

	var resp GetCategorisationsResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetCategorisationsResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize categorisations response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}

func (c *Client) GetBaseVariable(ctx context.Context, input GetBaseVariableInput) (GetBaseVariableResponse, error) {
	logData := log.Data{
		"method":          http.MethodGet,
		"population_type": input.PopulationType,
		"variable":        input.Variable,
	}

	urlPath := fmt.Sprintf("population-types/%s/dimensions/%s/base", input.PopulationType, input.Variable)

	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, urlPath, nil)
	if err != nil {
		return GetBaseVariableResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}

	clientlog.Do(ctx, "getting base variable", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetBaseVariableResponse{}, dperrors.New(
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
		return GetBaseVariableResponse{}, err
	}

	var resp GetBaseVariableResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetBaseVariableResponse{}, dperrors.New(
			errors.Wrap(err, "unable to deserialize base variable response"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}
