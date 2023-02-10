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

type GetMetaDataInput struct {
	AuthTokens
	PopulationType string
}

type GetMetadataResponse struct {
	PopulationType   string `json:"population-type"`
	DefaultDatasetID string `json:"default_dataset_id"`
	Edition          string `json:"edition"`
	Version          int    `json:"version"`
}

func (c *Client) GetMetadata(ctx context.Context, input GetMetaDataInput) (GetMetadataResponse, error) {
	logData := log.Data{
		"method": http.MethodGet,
	}
	uri := fmt.Sprintf("/population-types/%s/metadata", input.PopulationType)
	vals := url.Values{}
	req, err := c.createGetRequest(ctx, input.UserAuthToken, input.ServiceAuthToken, uri, vals)
	if err != nil {
		return GetMetadataResponse{}, dperrors.New(
			err,
			dperrors.StatusCode(err),
			logData,
		)
	}
	clientlog.Do(ctx, "getting metadata", service, req.URL.String(), logData)

	res, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return GetMetadataResponse{}, dperrors.New(
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
		return GetMetadataResponse{}, err
	}

	var metadata GetMetadataResponse
	if err := json.NewDecoder(res.Body).Decode(&metadata); err != nil {
		return GetMetadataResponse{}, dperrors.New(
			errors.Wrap(err, "failed to unmarshal response from GetMetadata"),
			http.StatusInternalServerError,
			logData,
		)
	}

	return metadata, nil
}
