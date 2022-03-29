package cantabular_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/filter"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testDownloadServiceToken   = "Download"
	testServiceAuthTokenHeader = "X-Florence-Token"
	testAuthTokenHeader        = "Authorization"
	ifMatch                    = "ea1e031b-3064-427d-8fed-4b35123213"
	newETag                    = "eb31e352f140b8a965d008f5505153bc6c4f5b48"
)

var req = cantabular.SubmitFilterRequest{
	FilterID: "ea1e031b-3064-427d-8fed-4b35c99bf1a3",
	Dimensions: []filter.DimensionOptions{{
		Items: []filter.DimensionOption{{
			DimensionOptionsURL: "http://some.url/city",
			Option:              "City",
		}},
		Count:      3,
		Offset:     0,
		Limit:      0,
		TotalCount: 3,
	}},
	PopulationType: "population-type",
}

var successfulResponse = cantabular.SubmitFilterResponse{
	InstanceID:       "instance-id",
	DimensionListUrl: "http://some.url/filter/filder-id/dimensions",
	FilterID:         "filter-id",
	Events:           nil,
	Dataset: cantabular.SFDataset{
		ID:      "dataset-id",
		Edition: "2022",
		Version: 1,
	},
	Links: cantabular.Links{
		Version: cantabular.Link{
			HREF: "http://some.url",
			ID:   "version-id",
		},
		Self: cantabular.Link{
			HREF: "http://some.url",
			ID:   "self-id",
		},
	},
	PopulationType: "population-type",
	Dimensions:     nil,
}

func Test_SubmitFilter(t *testing.T) {
	ctx := context.Background()
	Convey("Given a valid Submit Filter Request ", t, func() {
		Convey("when 'SubmitFilter' is called with the expected ifMatch value", func() {

			mockHttpClient, cantabularClient := mockedClient(newExpectedResponse(successfulResponse, http.StatusOK, newETag), nil)
			res, ETag, err := cantabularClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, fmt.Sprintf("cantabular.ext.host/filters/%s/submit", req.FilterID))
			})

			Convey("And the expected response is returned", func() {
				So(*res, ShouldResemble, successfulResponse)
			})

			Convey("And the expected ETag is empty", func() {
				So(ETag, ShouldEqual, newETag)
			})
		})

		Convey("when 'SubmitFilter' is called with an outdated ifMatch value", func() {
			var mockRespETagConflict = `{"message": "conflict: invalid ETag provided or filter has been updated"}`
			mockHttpClient, cantabularClient := mockedClient(newExpectedResponse(mockRespETagConflict, http.StatusConflict, ""), nil)
			res, ETag, err := cantabularClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then an error should be returned", func() {
				So(err.(filter.ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
				So(err.(filter.ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusConflict)
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, fmt.Sprintf("cantabular.ext.host/filters/%s/submit", req.FilterID))
			})

			Convey("And the expected response is returned", func() {
				So(res, ShouldBeNil)
			})

			Convey("And the expected ETag is empty", func() {
				So(ETag, ShouldEqual, "")
			})
		})

		Convey("when 'SubmitFilter' is called and the POST method returns an error", func() {
			mockHttpClient, cantabularClient := mockedClient(nil, errors.New("Something went wrong"))
			res, ETag, err := cantabularClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "failed to make request: Something went wrong")
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, fmt.Sprintf("cantabular.ext.host/filters/%s/submit", req.FilterID))
			})

			Convey("And the expected response is returned", func() {
				So(res, ShouldBeNil)
			})

			Convey("And the expected ETag is empty", func() {
				So(ETag, ShouldEqual, "")
			})
		})
	})
}

func newExpectedResponse(body interface{}, sc int, eTag string) *http.Response {
	b, _ := json.Marshal(body)

	expectedResponse := &http.Response{
		StatusCode: sc,
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		Header:     http.Header{},
	}
	expectedResponse.Header.Set("ETag", eTag)
	return expectedResponse
}

// newMockedClient creates a new cantabular client with a mocked response for post requests,
// according to the provided response string and status code.
func mockedClient(response *http.Response, err error) (*dphttp.ClienterMock, *cantabular.Client) {
	mockHttpClient := &dphttp.ClienterMock{
		PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
			return response, err
		},
	}

	cantabularClient := cantabular.NewClient(
		cantabular.Config{
			Host:       "cantabular.host",
			ExtApiHost: "cantabular.ext.host",
		},
		mockHttpClient,
		nil,
	)

	return mockHttpClient, cantabularClient
}
