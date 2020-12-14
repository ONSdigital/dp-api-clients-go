package filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testServiceToken         = "bar"
	testDownloadServiceToken = "baz"
	testUserAuthToken        = "grault"
	testCollectionID         = "garply"
	testHost                 = "http://localhost:8080"
)

var initialState = health.CreateCheckState(service)

// client with no retries, no backoff
var (
	client = &dphttp.Client{HTTPClient: &http.Client{}}
	ctx    = context.Background()
)

func checkResponseBase(httpClient *dphttp.ClienterMock, expectedMethod, expectedURI, serviceAuthToken string) {
	So(len(httpClient.DoCalls()), ShouldEqual, 1)
	So(httpClient.DoCalls()[0].Req.URL.RequestURI(), ShouldEqual, expectedURI)
	So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
	So(httpClient.DoCalls()[0].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+serviceAuthToken)
}

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		filterClient := newFilterClient(httpClient)

		check := initialState

		Convey("when filterClient.Checker is called", func() {
			err := filterClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 500,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		filterClient := newFilterClient(httpClient)
		check := initialState

		Convey("when filterClient.Checker is called", func() {
			err := filterClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 404,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		filterClient := newFilterClient(httpClient)
		check := initialState

		Convey("when filterClient.Checker is called", func() {
			err := filterClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 404)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 429,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		filterClient := newFilterClient(httpClient)
		check := initialState

		Convey("when filterClient.Checker is called", func() {
			err := filterClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, 429)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 200,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		filterClient := newFilterClient(httpClient)
		check := initialState

		Convey("when filterClient.Checker is called", func() {
			err := filterClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure(), ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_GetOutput(t *testing.T) {
	filterOutputID := "foo"
	filterOutputBody := `{"filter_id":"` + filterOutputID + `"}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a filter-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: filterOutputBody})
		model, err := mockedAPI.GetOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldBeNil)
		So(model, ShouldResemble, Model{FilterID: filterOutputID})
	})
}

func TestClient_UpdateFilterOutput(t *testing.T) {
	filterJobID := "filterID"
	model := Model{FilterID: filterJobID, InstanceID: "someInstance"}
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 500, Body: ""})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldNotBeNil)
	})

	Convey("When server returns 200 OK", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 200, Body: ""})
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldBeNil)
	})
}

func TestClient_GetDimension(t *testing.T) {
	filterOutputID := "foo"
	name := "corge"
	dimensionBody := `{
		"dimension_url": "www.ons.gov.uk",
		"name": "quuz",
		"options": ["corge"]}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		dim, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldBeNil)
		So(dim, ShouldResemble, Dimension{
			Name: "quuz",
			URI:  "www.ons.gov.uk",
		})
	})
}

func TestClient_GetDimensions(t *testing.T) {
	filterOutputID := "foo"
	dimensionBody := `[{
		"dimension_url": "www.ons.gov.uk",
		"name": "quuz",
		"options": ["corge"]}]`

	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})
		dims, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldBeNil)
		So(dims, ShouldResemble, []Dimension{
			Dimension{
				Name: "quuz",
				URI:  "www.ons.gov.uk",
			},
		})
	})
}

func TestClient_GetDimensionOptions(t *testing.T) {

	filterOutputID := "foo"
	dimensionBody := `{"items": [{"dimension_option_url":"quux","option": "quuz"}], "count":1, "offset":2, "limit": 10, "total_count": 3}`
	name := "corge"
	offset := 2
	limit := 10

	Convey("Given a 400 BadRequest response is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})

		Convey("then GetDimensionOptions returns the expected error", func() {
			_, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, offset, limit)
			So(err, ShouldResemble, &ErrInvalidFilterAPIResponse{
				ActualCode:   400,
				ExpectedCode: 200,
				URI:          fmt.Sprintf("%s/filters/%s/dimensions/%s/options?offset=%d&limit=%d", mockedAPI.hcCli.URL, filterOutputID, name, offset, limit),
			})
		})
	})

	Convey("Given a 500 InternalServerError is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)

		Convey("then GetDimensionOptions returns the expected error", func() {
			_, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, offset, limit)
			So(err, ShouldResemble, &ErrInvalidFilterAPIResponse{
				ActualCode:   500,
				ExpectedCode: 200,
				URI:          fmt.Sprintf("%s/filters/%s/dimensions/%s/options?offset=%d&limit=%d", mockedAPI.hcCli.URL, filterOutputID, name, offset, limit),
			})
		})
	})

	Convey("When a 200 OK status is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody})

		Convey("then GetDimensionOptions returns the expected Options", func() {
			opts, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, offset, limit)
			So(err, ShouldBeNil)
			So(opts, ShouldResemble, DimensionOptions{
				Items: []DimensionOption{
					{
						DimensionOptionsURL: "quux",
						Option:              "quuz",
					},
				},
				Count:      1,
				TotalCount: 3,
				Limit:      10,
				Offset:     2,
			})
		})

		Convey("then GetDimensionOptions returns the expected error when a negative offset is provided", func() {
			_, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, -1, limit)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
		})

		Convey("then GetDimensionOptions returns the expected error when a negative limit is provided", func() {
			_, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, offset, -1)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
		})
	})
}

func TestClient_CreateBlueprint(t *testing.T) {
	datasetID := "foo"
	edition := "quux"
	version := "1"
	names := []string{"quuz", "corge"}

	checkResponse := func(httpClient *dphttp.ClienterMock, expectedFilterID string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualVersion string
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedFilterID)
	}

	Convey("Given a valid Blueprint is returned", t, func() {

		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusCreated, 500, url + "/filters"}
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})
}

func TestClient_UpdateBlueprint(t *testing.T) {
	model := Model{
		FilterID:    "",
		InstanceID:  "",
		Links:       Links{},
		DatasetID:   "",
		Edition:     "",
		Version:     "",
		State:       "",
		Dimensions:  nil,
		Downloads:   nil,
		Events:      nil,
		IsPublished: false,
	}
	doSubmit := true

	checkResponse := func(httpClient *dphttp.ClienterMock, expectedModel Model) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualModel Model

		json.Unmarshal(actualBody, &actualModel)
		So(actualModel, ShouldResemble, expectedModel)
	}

	Convey("Given a valid blueprint update is given", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusOK, 500, url + "/filters/?submitted=" + strconv.FormatBool(doSubmit)}
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when CreateBlueprint is called", func() {
			bp, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, bp)
			})
		})
	})

}

func TestClient_AddDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"

	Convey("Given a valid dimension value is added", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValue is called", func() {
			err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValue is called", func() {
			err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name + "/options/filter-api"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusCreated, 500, uri}
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValue is called", func() {
			err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_RemoveDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"
	Convey("Given a dimension value is removed", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValue is called", func() {
			err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValue is called", func() {
			err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name + "/options/filter-api"
		mockInvalidStatusCodeError := ErrInvalidFilterAPIResponse{http.StatusNoContent, 500, uri}
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValue is called", func() {
			err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_AddDimension(t *testing.T) {
	filterID := "baz"
	name := "quz"

	Convey("Given a dimension is provided", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimension is called", func() {
			err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimension is called", func() {
			err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		mockInvalidStatusCodeError := errors.New("invalid status from filter api")
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimension is called", func() {
			err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

// utility func to validate request method, uri and body
func checkRequest(httpClient *dphttp.ClienterMock, callIndex int, expectedURI string, expectedPatchOp dprequest.PatchOp, expectedPatchValues []string) {
	So(httpClient.DoCalls()[callIndex].Req.URL.RequestURI(), ShouldEqual, expectedURI)
	So(httpClient.DoCalls()[callIndex].Req.Method, ShouldEqual, http.MethodPatch)
	So(httpClient.DoCalls()[callIndex].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+testServiceToken)
	expectedBody := []dprequest.Patch{
		{
			Op:    expectedPatchOp.String(),
			Path:  "/options/-",
			Value: expectedPatchValues,
		},
	}
	sentPayload, err := ioutil.ReadAll(httpClient.DoCalls()[callIndex].Req.Body)
	So(err, ShouldBeNil)
	var sentBody []dprequest.Patch
	err = json.Unmarshal(sentPayload, &sentBody)
	So(err, ShouldBeNil)
	So(sentBody, ShouldResemble, expectedBody)
}

func checkRequestTwoOps(httpClient *dphttp.ClienterMock, callIndex int, expectedURI string, expectedPatchOp1, expectedPatchOp2 dprequest.PatchOp, expectedPatchValues1, expectedPatchValues2 []string) {
	So(httpClient.DoCalls()[callIndex].Req.URL.RequestURI(), ShouldEqual, expectedURI)
	So(httpClient.DoCalls()[callIndex].Req.Method, ShouldEqual, http.MethodPatch)
	So(httpClient.DoCalls()[callIndex].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+testServiceToken)
	expectedBody := []dprequest.Patch{
		{
			Op:    expectedPatchOp1.String(),
			Path:  "/options/-",
			Value: expectedPatchValues1,
		},
		{
			Op:    expectedPatchOp2.String(),
			Path:  "/options/-",
			Value: expectedPatchValues2,
		},
	}
	sentPayload, err := ioutil.ReadAll(httpClient.DoCalls()[callIndex].Req.Body)
	So(err, ShouldBeNil)
	var sentBody []dprequest.Patch
	err = json.Unmarshal(sentPayload, &sentBody)
	So(err, ShouldBeNil)
	So(sentBody, ShouldResemble, expectedBody)
}

func TestClient_AddDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	batchSize := 5

	Convey("Given a dimension is provided", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValues is called, where total options are less than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl"}
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected PATCH body is generated and sent to the API", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				checkRequest(httpClient, 0, "/filters/"+filterID+"/dimensions/"+name, dprequest.OpAdd, options)
			})
		})

		Convey("when AddDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name

				So(len(httpClient.DoCalls()), ShouldEqual, 2)
				checkRequest(httpClient, 0, expectedURI, dprequest.OpAdd, []string{"abc", "def", "ghi", "jkl", "000"})
				checkRequest(httpClient, 1, expectedURI, dprequest.OpAdd, []string{"111", "222"})
			})
		})

		Convey("When AddDimensionValues is called with an empty list of options", func() {
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then no PATCH operation is sent", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValues is called, where total options are less than the batch size", func() {
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})

		Convey("when AddDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValues is called", func() {
			err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize)

			Convey("then the expected error is returned", func() {
				expectedErr := ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusInternalServerError,
					URI:          "http://localhost:8080/filters/baz/dimensions/quz",
				}
				So(err.Error(), ShouldResemble, expectedErr.Error())
			})
		})
	})
}

func TestClient_RemoveDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	batchSize := 5

	Convey("Given a dimension is provided", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValues is called", func() {
			options := []string{"abc", "def", "ghi", "jkl"}
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected URI and PATCH body is generated and sent to the API", func() {
				checkResponseBase(httpClient, http.MethodPatch, "/filters/"+filterID+"/dimensions/"+name, testServiceToken)
				Convey("The expected PATCH body is generated and sent to the API", func() {
					So(len(httpClient.DoCalls()), ShouldEqual, 1)
					checkRequest(httpClient, 0, "/filters/"+filterID+"/dimensions/"+name, dprequest.OpRemove, options)
				})
			})
		})

		Convey("when RemoveDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name

				So(len(httpClient.DoCalls()), ShouldEqual, 2)
				checkRequest(httpClient, 0, expectedURI, dprequest.OpRemove, []string{"abc", "def", "ghi", "jkl", "000"})
				checkRequest(httpClient, 1, expectedURI, dprequest.OpRemove, []string{"111", "222"})
			})
		})

		Convey("When RemoveDimensionValues is called with an empty list of options", func() {
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then no PATCH operation is sent", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValues is called", func() {
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})

		Convey("when RemoveDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValues is called", func() {
			err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize)

			Convey("then the expected error is returned", func() {
				expectedErr := ErrInvalidFilterAPIResponse{
					ExpectedCode: http.StatusOK,
					ActualCode:   http.StatusInternalServerError,
					URI:          "http://localhost:8080/filters/baz/dimensions/quz",
				}
				So(err.Error(), ShouldResemble, expectedErr.Error())
			})
		})
	})
}

func TestClient_PatchDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	batchSize := 5

	Convey("Given a dimension is provided", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when PatchDimensionValues is called", func() {
			optionsAdd := []string{"abc", "def"}
			optionsRemove := []string{"ghi", "jkl"}
			err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, optionsAdd, optionsRemove, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected URI and PATCH body is generated and sent to the API", func() {
				checkResponseBase(httpClient, http.MethodPatch, "/filters/"+filterID+"/dimensions/"+name, testServiceToken)
				Convey("The expected PATCH body is generated and sent to the API", func() {
					So(len(httpClient.DoCalls()), ShouldEqual, 1)
					checkRequestTwoOps(httpClient, 0, "/filters/"+filterID+"/dimensions/"+name, dprequest.OpAdd, dprequest.OpRemove, optionsAdd, optionsRemove)
				})
			})
		})

		Convey("when PatchDimensionValues is called, where total options are more than the batch size", func() {
			optionsAdd := []string{"abc", "def", "ghi"}
			optionsRemove := []string{"000", "111", "222"}
			err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, optionsAdd, optionsRemove, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name

				So(len(httpClient.DoCalls()), ShouldEqual, 2)
				checkRequest(httpClient, 0, expectedURI, dprequest.OpAdd, []string{"abc", "def", "ghi"})
				checkRequest(httpClient, 1, expectedURI, dprequest.OpRemove, []string{"000", "111", "222"})
			})
		})

		Convey("When PatchDimensionValues is called with an empty list of options", func() {
			err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, []string{}, batchSize)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then no PATCH operation is sent", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestClient_GetJobState(t *testing.T) {
	filterID := "foo"
	mockJobStateBody := `{
		"jobState": "www.ons.gov.uk"}`
	Convey("When a state is returned", t, func() {

		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: mockJobStateBody})
		_, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldBeNil)
	})
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		m, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldNotBeNil)
		So(m, ShouldResemble, Model{})
	})
}

func TestClient_SetDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	options := []string{"`quuz"}

	Convey("Given a valid dimension and filter", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when SetDimensionValues is called", func() {
			err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when SetDimensionValues is called", func() {
			err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		url := "http://localhost:8080"
		uri := url + "/filters/" + filterID + "/dimensions/" + name
		mockInvalidStatusCodeError := &ErrInvalidFilterAPIResponse{http.StatusCreated, http.StatusInternalServerError, uri}
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
		}, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when SetDimensionValues is called", func() {
			err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})
		})
	})
}

func TestClient_GetPreview(t *testing.T) {
	filterOutputID := "foo"
	previewBody := `{"somePreview":""}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetPreview(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetPreview(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When a preview is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: previewBody})
		p, err := mockedAPI.GetPreview(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldBeNil)
		So(p, ShouldResemble, Preview{})
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newFilterClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	filterClient := NewWithHealthClient(healthClient)
	return filterClient
}

func getMockfilterAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	return New(ts.URL)
}
