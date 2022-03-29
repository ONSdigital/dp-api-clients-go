package cantabular

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/clientlog"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	service     = "dp-cantabular-filter-flex-api"
	POST_METHOD = "POST"
)

type SubmitFilterRequest struct {
	FilterID       string                    `json:"filter_id"`
	Dimensions     []filter.DimensionOptions `json:"dimension_options,omitempty"`
	PopulationType string                    `json:"population_type"`
}

type Event struct {
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Name      string    `bson:"name"      json:"name"`
}

type SFDataset struct {
	ID      string `bson:"id"      json:"id"`
	Edition string `bson:"edition" json:"edition"`
	Version int    `bson:"version" json:"version"`
}

type Link struct {
	HREF string `bson:"href"           json:"href"`
	ID   string `bson:"id,omitempty"   json:"id,omitempty"`
}

type Links struct {
	Version Link `bson:"version" json:"version"`
	Self    Link `bson:"self"    json:"self"`
}

type SFDimension struct {
	Name         string   `bson:"name"          json:"name"`
	Options      []string `bson:"options"       json:"options"`
	DimensionURL string   `bson:"dimension_url" json:"dimension_url,omitempty"`
	IsAreaType   bool     `bson:"is_area_type"  json:"is_area_type"`
}

type SubmitFilterResponse struct {
	InstanceID       string        `json:"instance_id"`
	DimensionListUrl string        `json:"dimension_list_url"`
	FilterID         string        `json:"filter_id"`
	Events           []Event       `json:"events"`
	Dataset          SFDataset     `json:"dataset"`
	Links            Links         `json:"links"`
	PopulationType   string        `json:"population_type"`
	Dimensions       []SFDimension `json:"dimensions"`
}

// SubmitFilter function to submit the request to submit a filter for a cantabular dataset.
// Should POST to /filters/{filterid}/submit in dp-cantabular-filter-flex-api microservice.
func (c *Client) SubmitFilter(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceToken, ifMatch string, sfr SubmitFilterRequest) (*SubmitFilterResponse, string, error) {
	b, err := json.Marshal(sfr)
	if err != nil {
		return nil, "", err
	}

	uri := fmt.Sprintf("%s/filters/%s/submit", c.extApiHost, sfr.FilterID)

	clientlog.Do(ctx, "updating filter job", service, uri, log.Data{
		"method": POST_METHOD,
		"body":   string(b),
	})

	req, err := http.NewRequest(POST_METHOD, uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, "", err
	}

	if err = headers.SetAuthToken(req, userAuthToken); err != nil {
		return nil, "", fmt.Errorf("failed to set auth token: %w", err)
	}
	if err = headers.SetServiceAuthToken(req, serviceAuthToken); err != nil {
		return nil, "", fmt.Errorf("failed to set service auth token: %w", err)
	}
	if err = headers.SetDownloadServiceToken(req, downloadServiceToken); err != nil {
		return nil, "", fmt.Errorf("failed to set download service token: %w", err)
	}
	if err = headers.SetIfMatch(req, ifMatch); err != nil {
		return nil, "", fmt.Errorf("failed to set if match: %w", err)
	}

	buf := &bytes.Buffer{}
	resp, err := c.httpPost(ctx, uri, "application/json", buf)
	if err != nil {
		return nil, "", err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, "", filter.ErrInvalidFilterAPIResponse{ExpectedCode: http.StatusOK, ActualCode: resp.StatusCode, URI: uri}
	}

	eTag, err := headers.GetResponseETag(resp)
	if err != nil && err != headers.ErrHeaderNotFound {
		return nil, "", err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	var r *SubmitFilterResponse
	if err = json.Unmarshal(b, &r); err != nil {
		return nil, "", err
	}

	return r, eTag, nil
}
