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
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/filter"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/pkg/errors"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
)

const (
	testServiceToken         = "bar"
	testDownloadServiceToken = "baz"
	testUserAuthToken        = "grault"
	testCollectionID         = "garply"
	testHost                 = "http://localhost:8080"
	testETag                 = "1cf582ea0f9266de2686d6e246a243d15feacda3"
	testETag2                = "17f20b6965501e27adcd46c674f65eeb17abb7b9"
)

var initialState = health.CreateCheckState(service)

// client with no retries, no backoff
var (
	client = &dphttp.Client{HTTPClient: &http.Client{}}
	ctx    = context.Background()
)

// checkRequest validates request method, uri and headers
func checkRequest(httpClient *dphttp.ClienterMock, callIndex int, expectedMethod, expectedURI string, expectedIfMatch string) {
	So(httpClient.DoCalls()[callIndex].Req.URL.RequestURI(), ShouldEqual, expectedURI)
	So(httpClient.DoCalls()[callIndex].Req.Method, ShouldEqual, http.MethodPatch)
	So(httpClient.DoCalls()[callIndex].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+testServiceToken)
	actualIfMatch := httpClient.DoCalls()[callIndex].Req.Header.Get("If-Match")
	So(actualIfMatch, ShouldResemble, expectedIfMatch)
}

// getRequestPatchBody returns the patch request body sent with the provided httpClient in iteration callIndex
var getRequestPatchBody = func(httpClient *dphttp.ClienterMock, callIndex int) []dprequest.Patch {
	sentPayload, err := ioutil.ReadAll(httpClient.DoCalls()[callIndex].Req.Body)
	So(err, ShouldBeNil)
	var sentBody []dprequest.Patch
	err = json.Unmarshal(sentPayload, &sentBody)
	So(err, ShouldBeNil)
	return sentBody
}

var validateRequestPatches = func(httpClient *dphttp.ClienterMock, callIndex int, expectedPatches []dprequest.Patch) {
	sentPatches := getRequestPatchBody(httpClient, callIndex)
	So(len(sentPatches), ShouldEqual, len(expectedPatches))
	for i, patch := range expectedPatches {
		So(sentPatches[i].Op, ShouldEqual, patch.Op)
		So(sentPatches[i].Path, ShouldEqual, patch.Path)
		So(sentPatches[i].Value, ShouldResemble, patch.Value)
	}
}

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
	ETag       string
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

	Convey("When server error is returned in all attempts then the expected error is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in first attempt but 200 OK is returned in the first retry then the corresponding Output is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: filterOutputBody})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		model, err := mockedAPI.GetOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldBeNil)
		So(model, ShouldResemble, Model{FilterID: filterOutputID})
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

	Convey("When server error is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 500, Body: ""})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 200, Body: ""})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldBeNil)
	})

	Convey("When server returns 200 OK", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 200, Body: ""})
		err := mockedAPI.UpdateFilterOutput(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &model)
		So(err, ShouldBeNil)
	})
}

func TestClient_AddEvent(t *testing.T) {
	filterJobID := "filterID"
	event := Event{Type: EventFilterOutputCSVGenEnd, Time: time.Now()}
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "POST"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		err := mockedAPI.AddEvent(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &event)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "PPOST"},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 500, Body: ""})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		err := mockedAPI.AddEvent(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &event)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "POST"},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
			MockedHTTPResponse{StatusCode: 200, Body: ""})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		err := mockedAPI.AddEvent(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &event)
		So(err, ShouldBeNil)
	})

	Convey("When server returns 200 OK", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "POST"}, MockedHTTPResponse{StatusCode: 200, Body: ""})
		err := mockedAPI.AddEvent(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, filterJobID, &event)
		So(err, ShouldBeNil)
	})
}

func TestClient_GetDimension(t *testing.T) {
	filterOutputID := "foo"
	name := "corge"
	dimensionBody := `{
		"dimension_url": "www.ons.gov.uk",
		"name": "quuz",
		"is_area_type": false,
		"options": ["corge"]}`
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, _, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, _, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		dim, eTag, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldBeNil)
		So(dim, ShouldResemble, Dimension{
			Name:       "quuz",
			URI:        "www.ons.gov.uk",
			Options:    []string{"corge"},
			IsAreaType: boolToPtr(false),
		})
		So(eTag, ShouldResemble, testETag)
	})

	Convey("When a dimension-instance is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})
		dim, eTag, err := mockedAPI.GetDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name)
		So(err, ShouldBeNil)
		So(dim, ShouldResemble, Dimension{
			Name:       "quuz",
			URI:        "www.ons.gov.uk",
			Options:    []string{"corge"},
			IsAreaType: boolToPtr(false),
		})
		So(eTag, ShouldResemble, testETag)
	})
}

func TestClient_GetDimensions(t *testing.T) {
	filterOutputID := "foo"
	dimensionBody := `{
		"items": [
			{
				"dimension_url": "www.ons.gov.uk/dim1",
				"name": "DimensionOne",
				"options": ["one"],
				"is_area_type": false
			},
			{
				"dimension_url": "www.ons.gov.uk/dim2",
				"name": "DimensionTwo",
				"options": ["two"],
				"is_area_type": true
			}
		],
		"count": 2,
		"offset": 1,
		"limit": 20,
		"total_count": 3
	}`

	Convey("When bad request is returned then the expected ErrInvalidFilterAPIResponse is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, _, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, nil)
		So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
		So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusBadRequest)
		So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "/filters/foo/dimensions"), ShouldBeTrue)
	})

	Convey("When server error is returned in all attempts then the expected ErrInvalidFilterAPIResponse is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, _, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, nil)
		So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
		So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusInternalServerError)
		So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "/filters/foo/dimensions"), ShouldBeTrue)
	})

	Convey("When server error is returned in first attempt but 200 OK is returned in the first retry then the corresponding Dimensions struct is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		dims, eTag, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, nil)
		So(err, ShouldBeNil)
		So(dims, ShouldResemble, Dimensions{
			Items: []Dimension{
				{
					URI:        "www.ons.gov.uk/dim1",
					Name:       "DimensionOne",
					Options:    []string{"one"},
					IsAreaType: boolToPtr(false),
				},
				{
					URI:        "www.ons.gov.uk/dim2",
					Name:       "DimensionTwo",
					Options:    []string{"two"},
					IsAreaType: boolToPtr(true),
				},
			},
			Count:      2,
			Offset:     1,
			Limit:      20,
			TotalCount: 3,
		})
		So(eTag, ShouldResemble, testETag)
	})

	Convey("When a dimension-instance json is returned by the api then the corresponding Dimensions struct is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})

		Convey("Then a request with valid query parameterse returns the expected Dimensions struct", func() {
			q := QueryParams{Offset: 1, Limit: 20}
			dims, eTag, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, &q)
			So(err, ShouldBeNil)
			So(dims, ShouldResemble, Dimensions{
				Items: []Dimension{
					{
						URI:        "www.ons.gov.uk/dim1",
						Name:       "DimensionOne",
						Options:    []string{"one"},
						IsAreaType: boolToPtr(false),
					},
					{
						URI:        "www.ons.gov.uk/dim2",
						Name:       "DimensionTwo",
						Options:    []string{"two"},
						IsAreaType: boolToPtr(true),
					},
				},
				Count:      2,
				Offset:     1,
				Limit:      20,
				TotalCount: 3,
			})
			So(eTag, ShouldResemble, testETag)
		})

		Convey("Then a request with invalid offset query paratmers returns a validation error", func() {
			q := QueryParams{Offset: -1, Limit: 0}
			_, _, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, &q)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
		})

		Convey("Then a request with invalid limit query paratmers returns a validation error", func() {
			q := QueryParams{Offset: 0, Limit: -1}
			_, _, err := mockedAPI.GetDimensions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, &q)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
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
			_, _, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, nil)
			So(err, ShouldResemble, &ErrInvalidFilterAPIResponse{
				ActualCode:   400,
				ExpectedCode: 200,
				URI:          fmt.Sprintf("%s/filters/%s/dimensions/%s/options", mockedAPI.hcCli.URL, filterOutputID, name),
			})
		})
	})

	Convey("Given a 500 InternalServerError is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)

		Convey("then GetDimensionOptions returns the expected error", func() {
			_, _, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, nil)
			So(err, ShouldResemble, &ErrInvalidFilterAPIResponse{
				ActualCode:   500,
				ExpectedCode: 200,
				URI:          fmt.Sprintf("%s/filters/%s/dimensions/%s/options", mockedAPI.hcCli.URL, filterOutputID, name),
			})
		})
	})

	Convey("Given a 500 InternalServerError is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})
		mockedAPI.hcCli.Client.SetMaxRetries(2)

		Convey("then GetDimensionOptions returns the expected Options", func() {
			q := QueryParams{Offset: offset, Limit: limit}
			opts, eTag, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, &q)
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
			So(eTag, ShouldResemble, testETag)
		})
	})

	Convey("When a 200 OK status is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: dimensionBody, ETag: testETag})

		Convey("then GetDimensionOptions returns the expected Options", func() {
			q := QueryParams{Offset: offset, Limit: limit}
			opts, eTag, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, &q)
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
			So(eTag, ShouldResemble, testETag)
		})

		Convey("then GetDimensionOptions returns the expected error when a negative offset is provided", func() {
			q := QueryParams{Offset: -1, Limit: limit}
			_, _, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, &q)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
		})

		Convey("then GetDimensionOptions returns the expected error when a negative limit is provided", func() {
			q := QueryParams{Offset: offset, Limit: -1}
			_, _, err := mockedAPI.GetDimensionOptions(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, &q)
			So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
		})
	})
}

func TestClient_GetDimensionOptionsInBatches(t *testing.T) {

	filterOutputID := "foo"
	dimensionBody0 := `{"items": [
		{"dimension_option_url":"http://op1.co.uk", "option": "op1"},
		{"dimension_option_url":"http://op2.co.uk", "option": "op2"}
		], "offset": 0, "limit": 2, "count": 2, "total_count": 3}`
	dimensionBody1 := `{"items": [
		{"dimension_option_url":"http://op3.co.uk", "option": "op3"}
		], "offset": 2, "limit": 2, "count": 1, "total_count": 3}`
	name := "corge"
	batchSize := 2
	maxWorkers := 1

	Convey("Given a mocked batch processor", t, func() {

		// testProcess is a generic batch processor for testing
		processedBatches := []DimensionOptions{}
		processedETags := []string{}
		var testProcess DimensionOptionsBatchProcessor = func(batch DimensionOptions, eTag string) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			processedETags = append(processedETags, eTag)
			return false, nil
		}

		Convey("When 200 OK is returned in 2 consecutive calls, with the same eTag value", func() {

			// mockedAPI is a HTTP mock
			mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
				MockedHTTPResponse{StatusCode: 200, Body: dimensionBody0, ETag: testETag},
				MockedHTTPResponse{StatusCode: 200, Body: dimensionBody1, ETag: testETag},
			)

			Convey("Then GetDimensionOptionsInBatches succeeds and returns the accumulated items from all the batches along with the expected eTag", func() {
				opts, eTag, err := mockedAPI.GetDimensionOptionsInBatches(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, batchSize, maxWorkers)
				So(err, ShouldBeNil)
				So(opts, ShouldResemble, DimensionOptions{
					Items: []DimensionOption{
						{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
						{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
						{DimensionOptionsURL: "http://op3.co.uk", Option: "op3"},
					},
					Count:      3,
					TotalCount: 3,
					Limit:      0,
					Offset:     0,
				})
				So(eTag, ShouldResemble, testETag)
			})

			Convey("Then GetDimensionOptionsBatchProcess, with eTag validation enabled, calls the batchProcessor function twice, with the expected baches and ETags", func() {
				eTag, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, true)
				So(err, ShouldBeNil)
				So(processedBatches, ShouldResemble, []DimensionOptions{
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
							{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
						},
						Count:      2,
						TotalCount: 3,
						Limit:      2,
						Offset:     0,
					},
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op3.co.uk", Option: "op3"},
						},
						Count:      1,
						TotalCount: 3,
						Limit:      2,
						Offset:     2,
					},
				})
				So(processedETags, ShouldResemble, []string{testETag, testETag})
				So(eTag, ShouldResemble, testETag)
			})

			Convey("Then GetDimensionOptionsBatchProcess, with eTag validation disabled, calls the batchProcessor function twice, with the expected baches and ETags", func() {
				eTag, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, false)
				So(err, ShouldBeNil)
				So(processedBatches, ShouldResemble, []DimensionOptions{
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
							{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
						},
						Count:      2,
						TotalCount: 3,
						Limit:      2,
						Offset:     0,
					},
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op3.co.uk", Option: "op3"},
						},
						Count:      1,
						TotalCount: 3,
						Limit:      2,
						Offset:     2,
					},
				})
				So(processedETags, ShouldResemble, []string{testETag, testETag})
				So(eTag, ShouldResemble, testETag)
			})
		})

		Convey("When 200 OK is returned in 2 consecutive calls, with different eTag values", func() {

			// mockedAPI is a HTTP mock
			mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
				MockedHTTPResponse{StatusCode: 200, Body: dimensionBody0, ETag: testETag},
				MockedHTTPResponse{StatusCode: 200, Body: dimensionBody1, ETag: testETag2},
			)

			Convey("Then GetDimensionOptionsInBatches fails due to the eTag mismatch between batches", func() {
				_, _, err := mockedAPI.GetDimensionOptionsInBatches(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, batchSize, maxWorkers)
				So(err, ShouldResemble, ErrBatchETagMismatch)
			})

			Convey("Then GetDimensionOptionsBatchProcess, with eTag validation enabled, fails due to the eTag mismatch between batches, and only the first batch is processed", func() {
				_, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, true)
				So(err, ShouldResemble, ErrBatchETagMismatch)
				So(processedBatches, ShouldResemble, []DimensionOptions{
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
							{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
						},
						Count:      2,
						TotalCount: 3,
						Limit:      2,
						Offset:     0,
					},
				})
				So(processedETags, ShouldResemble, []string{testETag})
			})

			Convey("Then GetDimensionOptionsBatchProcess, with eTag validation disabled, calls the batchProcessor function twice, with the expected baches and ETags", func() {
				eTag, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, false)
				So(err, ShouldBeNil)
				So(processedBatches, ShouldResemble, []DimensionOptions{
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
							{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
						},
						Count:      2,
						TotalCount: 3,
						Limit:      2,
						Offset:     0,
					},
					{
						Items: []DimensionOption{
							{DimensionOptionsURL: "http://op3.co.uk", Option: "op3"},
						},
						Count:      1,
						TotalCount: 3,
						Limit:      2,
						Offset:     2,
					},
				})
				So(processedETags, ShouldResemble, []string{testETag, testETag2})
				So(eTag, ShouldResemble, testETag2)
			})
		})
	})

	Convey("When a 400 error status is returned in the first call", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 400, Body: ""})

		// testProcess is a generic batch processor for testing
		processedBatches := []DimensionOptions{}
		var testProcess DimensionOptionsBatchProcessor = func(batch DimensionOptions, batchETag string) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetDimensionOptionsInBatches fails with the expected error and the process is aborted", func() {
			_, _, err := mockedAPI.GetDimensionOptionsInBatches(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, batchSize, maxWorkers)
			So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
			So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusBadRequest)
			So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "filters/foo/dimensions/corge/options?offset=0&limit=2"), ShouldBeTrue)
		})

		Convey("then GetDimensionOptionsBatchProcess fails with the expected error and doesn't call the batchProcessor", func() {
			_, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, true)
			So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
			So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusBadRequest)
			So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "filters/foo/dimensions/corge/options?offset=0&limit=2"), ShouldBeTrue)
			So(processedBatches, ShouldResemble, []DimensionOptions{})
		})
	})

	Convey("When a 200 error status is returned in the first call and 400 error is returned in the second call", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 200, Body: dimensionBody0},
			MockedHTTPResponse{StatusCode: 400, Body: ""})

		// testProcess is a generic batch processor for testing
		processedBatches := []DimensionOptions{}
		processedETags := []string{}
		var testProcess DimensionOptionsBatchProcessor = func(batch DimensionOptions, batchEtag string) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			processedETags = append(processedETags, batchEtag)
			return false, nil
		}

		Convey("Then GetDimensionOptionsInBatches fails with the expected error", func() {
			_, _, err := mockedAPI.GetDimensionOptionsInBatches(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, batchSize, maxWorkers)
			So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
			So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusBadRequest)
			So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "filters/foo/dimensions/corge/options?offset=2&limit=2"), ShouldBeTrue)
		})

		Convey("then GetDimensionOptionsBatchProcess fails with the expected error and calls the batchProcessor for the first batch only", func() {
			_, err := mockedAPI.GetDimensionOptionsBatchProcess(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterOutputID, name, testProcess, batchSize, maxWorkers, true)
			So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusOK)
			So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusBadRequest)
			So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "filters/foo/dimensions/corge/options?offset=2&limit=2"), ShouldBeTrue)
			So(processedBatches, ShouldResemble, []DimensionOptions{
				{
					Items: []DimensionOption{
						{DimensionOptionsURL: "http://op1.co.uk", Option: "op1"},
						{DimensionOptionsURL: "http://op2.co.uk", Option: "op2"},
					},
					Count:      2,
					TotalCount: 3,
					Limit:      2,
					Offset:     0,
				},
			})
		})
	})
}

func TestClient_CreateFlexBlueprint(t *testing.T) {
	datasetID := "foo"
	edition := "quux"
	version := "1"
	names := []string{"quuz", "corge"}
	population_type := "Teaching-Dataset"

	expectedRequest := createFlexBlueprintRequest{
		Dataset: Dataset{DatasetID: "foo", Edition: "quux", Version: 1},
		Dimensions: []ModelDimension{
			{Name: "quuz"},
			{Name: "corge"},
		},
		PopulationType: "Teaching-Dataset",
	}

	checkRequest := func(httpClient *dphttp.ClienterMock) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualVersion createFlexBlueprintRequest
		err := json.Unmarshal(actualBody, &actualVersion)
		So(err, ShouldBeNil)
		So(actualVersion, ShouldResemble, expectedRequest)
	}

	Convey("Given a valid Blueprint is returned", t, func() {
		expectedFilterId := "the-filter-id"
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":"the-filter-id"}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", testETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			bp, eTag, err := filterClient.CreateFlexibleBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names, population_type)

			Convey("then the expectedRequest eTag is returned, with no error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, testETag)
			})

			Convey("and dphttp client is called one time with the expectedRequest parameters", func() {
				checkRequest(httpClient)
			})

			Convey("and the response's filter id is correctly unmarshalled", func() {
				So(bp, ShouldEqual, expectedFilterId)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			_, _, err := filterClient.CreateFlexibleBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names, population_type)

			Convey("then the expectedRequest error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
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

		Convey("when createBlueprint is called", func() {
			_, _, err := filterClient.CreateFlexibleBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names, population_type)

			Convey("then the expectedRequest error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})
		})
	})
}

func TestClient_CreateBlueprint(t *testing.T) {
	datasetID := "foo"
	edition := "quux"
	version := "1"
	names := []string{"quuz", "corge"}

	expectedRequest := createBlueprint{
		Dataset: Dataset{DatasetID: "foo", Edition: "quux", Version: 1},
		Dimensions: []ModelDimension{
			{Name: "quuz"},
			{Name: "corge"},
		},
	}

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedFilterID string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualVersion createBlueprint
		err := json.Unmarshal(actualBody, &actualVersion)
		So(err, ShouldBeNil)
		So(actualVersion, ShouldResemble, expectedRequest)
		So(actualVersion.FilterID, ShouldResemble, expectedFilterID)
	}

	Convey("Given a valid Blueprint is returned", t, func() {
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", testETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			bp, eTag, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then the expectedRequest eTag is returned, with no error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, testETag)
			})

			Convey("and dphttp client is called one time with the expectedRequest parameters", func() {
				checkRequest(httpClient, bp)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then the expectedRequest error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expectedRequest parameters", func() {
				checkRequest(httpClient, bp)
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

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.CreateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, datasetID, edition, version, names)

			Convey("then the expectedRequest error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expectedRequest parameters", func() {
				checkRequest(httpClient, bp)
			})
		})
	})
}

func Test_SubmitFilter(t *testing.T) {
	testDownloadServiceToken := "Download"
	testServiceAuthTokenHeader := "X-Florence-Token"
	testAuthTokenHeader := "Authorization"
	ifMatch := "ea1e031b-3064-427d-8fed-4b35123213"
	newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"

	ctx := context.Background()

	var req = SubmitFilterRequest{
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

	var successfulResponse = SubmitFilterResponse{
		InstanceID:       "instance-id",
		DimensionListUrl: "http://some.url/filter/filder-id/dimensions",
		FilterID:         "filter-id",
		Events:           nil,
		Dataset: Dataset{
			DatasetID: "dataset-id",
			Edition:   "2022",
			Version:   1,
		},
		Links: Links{
			Version: Link{
				HRef: "http://some.url",
				ID:   "version-id",
			},
		},
		PopulationType: "population-type",
		Dimensions:     nil,
	}

	var newExpectedResponse = func(body interface{}, sc int, eTag string) *http.Response {
		b, _ := json.Marshal(body)

		expectedResponse := &http.Response{
			StatusCode: sc,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     http.Header{},
		}
		expectedResponse.Header.Set("ETag", eTag)
		return expectedResponse
	}

	Convey("Given a valid Submit Filter Request ", t, func() {
		Convey("when 'SubmitFilter' is called with the expected ifMatch value", func() {
			httpClient := newMockHTTPClient(newExpectedResponse(successfulResponse, http.StatusOK, newETag), nil)
			filterClient := newFilterClient(httpClient)
			res, ETag, err := filterClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, fmt.Sprintf("%s/filters/%s/submit", filterClient.hcCli.URL, req.FilterID))
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

			httpClient := newMockHTTPClient(newExpectedResponse(mockRespETagConflict, http.StatusConflict, ""), nil)
			filterClient := newFilterClient(httpClient)
			res, ETag, err := filterClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then an error should be returned", func() {
				So(err.(*dperrors.Error).Code(), ShouldEqual, http.StatusConflict)
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, fmt.Sprintf("%s/filters/%s/submit", filterClient.hcCli.URL, req.FilterID))
			})

			Convey("And the expected response is returned", func() {
				So(res, ShouldBeNil)
			})

			Convey("And the expected ETag is empty", func() {
				So(ETag, ShouldEqual, "")
			})
		})

		Convey("when 'SubmitFilter' is called and the POST method returns an error", func() {
			mockError := errors.New("Something went wrong")
			httpClient := newMockHTTPClient(nil, mockError)
			filterClient := newFilterClient(httpClient)
			res, ETag, err := filterClient.SubmitFilter(ctx, testAuthTokenHeader, testServiceAuthTokenHeader, testDownloadServiceToken, ifMatch, req)

			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "failed to create submit request: Something went wrong")
			})

			Convey("And the expected query is posted to cantabular filter-flex-api", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, fmt.Sprintf("%s/filters/%s/submit", filterClient.hcCli.URL, req.FilterID))
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

func TestClient_UpdateBlueprint(t *testing.T) {
	model := Model{
		FilterID:    "",
		InstanceID:  "",
		Links:       Links{},
		Dataset:     Dataset{},
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

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedModel Model, expectedIfMatch string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualModel Model

		err := json.Unmarshal(actualBody, &actualModel)
		So(err, ShouldBeNil)
		So(actualModel, ShouldResemble, expectedModel)

		actualIfMatch := httpClient.DoCalls()[0].Req.Header.Get("If-Match")
		So(actualIfMatch, ShouldResemble, expectedIfMatch)
	}

	Convey("Given a valid blueprint update is given", t, func() {
		newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when UpdateBlueprint is called with the expected ifMatch value", func() {
			bp, eTag, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
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

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.UpdateBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
			})
		})
	})
}

func TestClient_UpdateFlexBlueprint(t *testing.T) {
	model := Model{
		FilterID:       "",
		InstanceID:     "",
		Links:          Links{},
		Dataset:        Dataset{},
		DatasetID:      "",
		Edition:        "",
		Version:        "",
		State:          "",
		Dimensions:     nil,
		Downloads:      nil,
		Events:         nil,
		IsPublished:    false,
		PopulationType: "",
	}
	doSubmit := true

	populationType := "population-type"

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedModel Model, expectedIfMatch string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
		var actualModel Model

		err := json.Unmarshal(actualBody, &actualModel)
		So(err, ShouldBeNil)
		So(actualModel, ShouldResemble, expectedModel)

		actualIfMatch := httpClient.DoCalls()[0].Req.Header.Get("If-Match")
		So(actualIfMatch, ShouldResemble, expectedIfMatch)
	}

	Convey("Given a valid blueprint update is given", t, func() {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		httpClient := newMockHTTPClient(r, nil)
		filterClient := newFilterClient(httpClient)
		bp, _, err := filterClient.UpdateFlexBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, populationType, testETag)

		model.PopulationType = "population-type"

		Convey("then the model should be returned with the updated PopulationType", func() {
			So(err, ShouldBeNil)
			So(bp, ShouldResemble, model)
		})

	})

	Convey("Given a valid blueprint update is given", t, func() {
		newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"
		r := &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when UpdateFlexBlueprint is called with the expected ifMatch value", func() {
			bp, eTag, err := filterClient.UpdateFlexBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, populationType, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")

		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.UpdateFlexBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, populationType, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
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

		Convey("when createBlueprint is called", func() {
			bp, _, err := filterClient.UpdateFlexBlueprint(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, model, doSubmit, populationType, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkRequest(httpClient, bp, testETag)
			})
		})
	})
}

func TestClient_AddDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"
	newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedIfMatch string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)
		actualIfMatch := httpClient.DoCalls()[0].Req.Header.Get("If-Match")
		So(actualIfMatch, ShouldResemble, expectedIfMatch)
	}

	Convey("Given a valid dimension value is added", t, func() {
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValue is called", func() {
			eTag, err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})

			Convey("then the expected ifMatch value is sent", func() {
				checkRequest(httpClient, testETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValue is called", func() {
			_, err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

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
			_, err := filterClient.AddDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})
		})
	})
}

func TestClient_RemoveDimensionValue(t *testing.T) {
	filterID := "baz"
	name := "quz"
	newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedIfMatch string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)
		actualIfMatch := httpClient.DoCalls()[0].Req.Header.Get("If-Match")
		So(actualIfMatch, ShouldResemble, expectedIfMatch)
	}

	Convey("Given a dimension value is removed", t, func() {
		r := &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValue is called", func() {
			eTag, err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})

			Convey("then the expected ifMatch value is sent", func() {
				checkRequest(httpClient, testETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValue is called", func() {
			_, err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

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
			_, err := filterClient.RemoveDimensionValue(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, service, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockInvalidStatusCodeError.Error())
			})

		})
	})
}

func TestClient_AddDimension(t *testing.T) {
	filterID := "baz"
	name := "quz"
	newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"

	checkRequest := func(httpClient *dphttp.ClienterMock, expectedIfMatch string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)
		actualIfMatch := httpClient.DoCalls()[0].Req.Header.Get("If-Match")
		So(actualIfMatch, ShouldResemble, expectedIfMatch)
	}

	Convey("Given a dimension is provided", t, func() {
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimension is called", func() {
			eTag, err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, testETag)

			Convey("then the expected eTag is returned without returned", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})

			Convey("then the expected ifMatch value is sent", func() {
				checkRequest(httpClient, testETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimension is called", func() {
			_, err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, testETag)

			Convey("then the expected error is returned", func() {
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

		Convey("when AddDimension is called", func() {
			_, err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, testETag)

			Convey("then the expected error is returned", func() {
				So(err.(*ErrInvalidFilterAPIResponse).ExpectedCode, ShouldEqual, http.StatusCreated)
				So(err.(*ErrInvalidFilterAPIResponse).ActualCode, ShouldEqual, http.StatusInternalServerError)
				So(strings.HasSuffix(err.(*ErrInvalidFilterAPIResponse).URI, "http://localhost:8080/filters/baz/dimensions/quz"), ShouldBeTrue)
			})
		})
	})
}

func TestClient_AddFlexDimension(t *testing.T) {
	const filterID = "baz"
	const name = "quz"
	const isAreaType = true
	const newETag = "eb31e352f140b8a965d008f5505153bc6c4f5b48"
	options := []string{"test"}

	Convey("Given a dimension is provided", t, func() {
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Header:     http.Header{"Etag": []string{newETag}},
		}

		httpClient := newMockHTTPClient(r, nil)
		filterClient := newFilterClient(httpClient)

		Convey("when AddFlexDimension is called", func() {
			eTag, err := filterClient.AddFlexDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, isAreaType, testETag)
			calls := httpClient.DoCalls()

			Convey("then the API should have been called once", func() {
				So(calls, ShouldHaveLength, 1)
			})

			Convey("then the correct headers should be sent", func() {
				headers := calls[0].Req.Header

				Convey("then the expected ifMatch value is sent", func() {
					So(headers.Get("If-Match"), ShouldEqual, testETag)
				})

				Convey("then the collection ID header is sent", func() {
					So(headers.Get("Collection-Id"), ShouldEqual, testCollectionID)
				})

				Convey("then the auth token is sent", func() {
					So(headers.Get("X-Florence-Token"), ShouldEqual, testUserAuthToken)
				})

				Convey("then the service token is sent", func() {
					So(headers.Get("Authorization"), ShouldEqual, fmt.Sprintf("Bearer %s", testServiceToken))
				})
			})

			Convey("then the request URL should contain the filter ID and name", func() {
				calledPath := calls[0].Req.URL.Path
				So(calledPath, ShouldEqual, fmt.Sprintf("/filters/%s/dimensions", filterID))
			})

			Convey("then the request body contains the passed dimension data", func() {
				var expectedDimension createFlexDimensionRequest
				err = json.NewDecoder(calls[0].Req.Body).Decode(&expectedDimension)
				So(err, ShouldBeNil)

				So(expectedDimension.Name, ShouldEqual, name)
				So(expectedDimension.IsAreaType, ShouldEqual, isAreaType)
				So(expectedDimension.Options, ShouldResemble, options)
			})

			Convey("then the expected eTag is returned", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when AddFlexDimension is called", func() {
			_, err := filterClient.AddFlexDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, nil, false, testETag)

			Convey("then the error contains the original client error", func() {
				So(errors.Is(err, mockErr), ShouldBeTrue)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		filterClient := newFilterClient(httpClient)

		Convey("when AddFlexDimension is called", func() {
			_, err := filterClient.AddDimension(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, testETag)

			Convey("then the error contains the original response error", func() {
				var filterErr *ErrInvalidFilterAPIResponse
				ok := errors.As(err, &filterErr)
				So(ok, ShouldBeTrue)

				So(filterErr.ExpectedCode, ShouldEqual, http.StatusCreated)
				So(filterErr.ActualCode, ShouldEqual, http.StatusInternalServerError)
				So(filterErr.URI, ShouldEqual, "http://localhost:8080/filters/baz/dimensions/quz")
			})
		})
	})
}

func TestClient_AddDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	batchSize := 5
	newETags := []string{
		"eb31e352f140b8a965d008f5505153bc6c4f5b48",
		"84798def3a75c8783b09e946d2fbf85e8a1dcce5"}

	Convey("Given a dimension is provided", t, func() {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
		}
		httpClient := newMockHTTPClient(r, nil)
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			r.Header.Set("ETag", newETags[len(httpClient.DoCalls())-1])
			return r, nil
		}

		filterClient := newFilterClient(httpClient)

		Convey("when AddDimensionValues is called, where total options are less than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl"}
			eTag, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

			Convey("then the expected eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[0])
			})

			Convey("The expected PATCH body is generated and sent to the API along with the expected If-Match header", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				checkRequest(httpClient, 0, http.MethodPatch, "/filters/"+filterID+"/dimensions/"+name, testETag)
				expectedPatches := []dprequest.Patch{
					{Op: dprequest.OpAdd.String(), Path: "/options/-", Value: []interface{}{"abc", "def", "ghi", "jkl"}},
				}
				validateRequestPatches(httpClient, 0, expectedPatches)
			})
		})

		Convey("when AddDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			eTag, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

			Convey("then the expected latest eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[1])
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name
				So(len(httpClient.DoCalls()), ShouldEqual, 2)

				checkRequest(httpClient, 0, http.MethodPatch, expectedURI, testETag)
				checkRequest(httpClient, 1, http.MethodPatch, expectedURI, newETags[0])
				validateRequestPatches(httpClient, 0, []dprequest.Patch{
					{Op: dprequest.OpAdd.String(), Path: "/options/-", Value: []interface{}{"abc", "def", "ghi", "jkl", "000"}},
				})
				validateRequestPatches(httpClient, 1, []dprequest.Patch{
					{Op: dprequest.OpAdd.String(), Path: "/options/-", Value: []interface{}{"111", "222"}},
				})
			})
		})

		Convey("When AddDimensionValues is called with an empty list of options", func() {
			eTag, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, batchSize, testETag)

			Convey("then the provided eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, testETag)
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
			_, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize, testETag)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})

		Convey("when AddDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			_, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

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
			_, err := filterClient.AddDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize, testETag)

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
	newETags := []string{
		"eb31e352f140b8a965d008f5505153bc6c4f5b48",
		"84798def3a75c8783b09e946d2fbf85e8a1dcce5"}

	Convey("Given a dimension is provided", t, func() {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
		}
		httpClient := newMockHTTPClient(r, nil)
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			r.Header.Set("ETag", newETags[len(httpClient.DoCalls())-1])
			return r, nil
		}

		filterClient := newFilterClient(httpClient)

		Convey("when RemoveDimensionValues is called", func() {
			options := []string{"abc", "def", "ghi", "jkl"}
			eTag, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[0])
			})

			Convey("The expected URI and PATCH body is generated and sent to the API", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				checkRequest(httpClient, 0, http.MethodPatch, "/filters/"+filterID+"/dimensions/"+name, testETag)
				validateRequestPatches(httpClient, 0, []dprequest.Patch{
					{Op: dprequest.OpRemove.String(), Path: "/options/-", Value: []interface{}{"abc", "def", "ghi", "jkl"}},
				})
			})
		})

		Convey("when RemoveDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			eTag, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

			Convey("then the latest eTag, obtained from the last call in the batch, is returned", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[1])
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches, each with its expected ifMatch value", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name
				So(len(httpClient.DoCalls()), ShouldEqual, 2)
				checkRequest(httpClient, 0, http.MethodPatch, expectedURI, testETag)
				checkRequest(httpClient, 1, http.MethodPatch, expectedURI, newETags[0])

				validateRequestPatches(httpClient, 0, []dprequest.Patch{
					{Op: dprequest.OpRemove.String(), Path: "/options/-", Value: []interface{}{"abc", "def", "ghi", "jkl", "000"}},
				})
				validateRequestPatches(httpClient, 1, []dprequest.Patch{
					{Op: dprequest.OpRemove.String(), Path: "/options/-", Value: []interface{}{"111", "222"}},
				})
			})
		})

		Convey("When RemoveDimensionValues is called with an empty list of options", func() {
			eTag, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, batchSize, testETag)

			Convey("then the original eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, testETag)
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
			_, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize, testETag)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
			})
		})

		Convey("when RemoveDimensionValues is called, where total options are more than the batch size", func() {
			options := []string{"abc", "def", "ghi", "jkl", "000", "111", "222"}
			_, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, batchSize, testETag)

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
			_, err := filterClient.RemoveDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{"abc"}, batchSize, testETag)

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
	newETags := []string{
		"eb31e352f140b8a965d008f5505153bc6c4f5b48",
		"84798def3a75c8783b09e946d2fbf85e8a1dcce5"}

	Convey("Given a dimension is provided", t, func() {
		r := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
		}
		httpClient := newMockHTTPClient(r, nil)
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			r.Header.Set("ETag", newETags[len(httpClient.DoCalls())-1])
			return r, nil
		}

		filterClient := newFilterClient(httpClient)

		Convey("when PatchDimensionValues is called", func() {
			optionsAdd := []string{"abc", "def"}
			optionsRemove := []string{"ghi", "jkl"}
			eTag, err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, optionsAdd, optionsRemove, batchSize, testETag)

			Convey("then the expected eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[0])
			})

			Convey("The expected URI and PATCH body is generated and sent to the API", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				checkRequest(httpClient, 0, http.MethodPatch, "/filters/"+filterID+"/dimensions/"+name, testETag)
				validateRequestPatches(httpClient, 0, []dprequest.Patch{
					{Op: dprequest.OpAdd.String(), Path: "/options/-", Value: []interface{}{"abc", "def"}},
					{Op: dprequest.OpRemove.String(), Path: "/options/-", Value: []interface{}{"ghi", "jkl"}},
				})

			})
		})

		Convey("when PatchDimensionValues is called, where total options are more than the batch size", func() {
			optionsAdd := []string{"abc", "def", "ghi"}
			optionsRemove := []string{"000", "111", "222"}
			eTag, err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, optionsAdd, optionsRemove, batchSize, testETag)

			Convey("then the latest eTag, obtained from the last call in the batch, is returned", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETags[1])
			})

			Convey("The expected PATCH body is generated and sent to the API in 2 batches, each with its expected ifMatch value", func() {
				expectedURI := "/filters/" + filterID + "/dimensions/" + name
				So(len(httpClient.DoCalls()), ShouldEqual, 2)

				checkRequest(httpClient, 0, http.MethodPatch, expectedURI, testETag)
				checkRequest(httpClient, 1, http.MethodPatch, expectedURI, newETags[0])
				validateRequestPatches(httpClient, 0, []dprequest.Patch{
					{Op: dprequest.OpAdd.String(), Path: "/options/-", Value: []interface{}{"abc", "def", "ghi"}},
				})
				validateRequestPatches(httpClient, 1, []dprequest.Patch{
					{Op: dprequest.OpRemove.String(), Path: "/options/-", Value: []interface{}{"000", "111", "222"}},
				})

			})
		})

		Convey("When PatchDimensionValues is called with an empty list of options", func() {
			eTag, err := filterClient.PatchDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, []string{}, []string{}, batchSize, testETag)

			Convey("then the original eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, testETag)
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
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 200, Body: mockJobStateBody, ETag: testETag})
		_, eTag, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldBeNil)
		So(eTag, ShouldResemble, testETag)
	})

	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, _, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		m, _, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldNotBeNil)
		So(m, ShouldResemble, Model{})
	})

	Convey("When server error is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: mockJobStateBody, ETag: testETag})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, eTag, err := mockedAPI.GetJobState(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterID)
		So(err, ShouldBeNil)
		So(eTag, ShouldResemble, testETag)
	})
}

func TestClient_SetDimensionValues(t *testing.T) {
	filterID := "baz"
	name := "quz"
	options := []string{"`quuz"}
	newETag := "eb31e352f140b8a965d008f5505153bc6c4f5b48"

	Convey("Given a valid dimension and filter", t, func() {
		r := &http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"filter_id":""}`))),
			Header:     http.Header{},
		}
		r.Header.Set("ETag", newETag)
		httpClient := newMockHTTPClient(r, nil)

		filterClient := newFilterClient(httpClient)

		Convey("when SetDimensionValues is called", func() {
			eTag, err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, testETag)

			Convey("then the new eTag is returned without error", func() {
				So(err, ShouldBeNil)
				So(eTag, ShouldResemble, newETag)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("foo")
		httpClient := newMockHTTPClient(nil, mockErr)

		filterClient := newFilterClient(httpClient)

		Convey("when SetDimensionValues is called", func() {
			_, err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, testETag)

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
			_, err := filterClient.SetDimensionValues(ctx, testUserAuthToken, testServiceToken, testCollectionID, filterID, name, options, testETag)

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

	Convey("When server error is returned in all attempts", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetPreview(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldNotBeNil)
	})

	Convey("When server error is returned in the first attempt but 200 OK is returned in the retry", t, func() {
		mockedAPI := getMockfilterAPI(http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "qux"},
			MockedHTTPResponse{StatusCode: 200, Body: previewBody})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		p, err := mockedAPI.GetPreview(ctx, testUserAuthToken, testServiceToken, testDownloadServiceToken, testCollectionID, filterOutputID)
		So(err, ShouldBeNil)
		So(p, ShouldResemble, Preview{})
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

func getMockfilterAPI(expectRequest http.Request, mockedHTTPResponse ...MockedHTTPResponse) *Client {
	numCall := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.Header().Set("ETag", mockedHTTPResponse[numCall].ETag)
		w.WriteHeader(mockedHTTPResponse[numCall].StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse[numCall].Body)
		numCall++
	}))
	return New(ts.URL)
}

func boolToPtr(val bool) *bool {
	return &val
}
