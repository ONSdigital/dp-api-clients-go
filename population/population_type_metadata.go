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

// GetPopulationTypeMetadataInput is a model with auth token and population type
type GetPopulationTypeMetadataInput struct {
	AuthTokens
	PopulationType string
}

// GetPopulationTypeMetadataResponse model with contain the metadata for poulation type
type GetPopulationTypeMetadataResponse struct {
	PopulationType   string `json:"population_type"`
	DefaultDatasetID string `json:"default_dataset_id"`
	Edition          string `json:"edition"`
	Version          int    `json:"version"`
}

func (c *Client) GetPopulationTypeMetadata(ctx context.Context, input GetPopulationTypeMetadataInput) (GetPopulationTypeMetadataResponse, error) {
	logData := log.Data{
		"method": http.MethodGet,
	}
	uri := fmt.Sprintf("/population-types/%s/metadata", input.PopulationType)
	vals := url.Values{}
	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, uri, vals)
	if err != nil {
		return GetPopulationTypeMetadataResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}
	clientlog.Do(ctx, "getting population type metadata", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetPopulationTypeMetadataResponse{}, dperrors.New(
			errors.Wrap(err, "failed to get response from Population types API (GetMetadata)"),
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
		return GetPopulationTypeMetadataResponse{}, err
	}

	var resp GetPopulationTypeMetadataResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return GetPopulationTypeMetadataResponse{}, dperrors.New(
			errors.Wrap(err, "failed to unmarshal response from GetMetadata"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return resp, nil
}
