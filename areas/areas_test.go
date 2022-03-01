package areas

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	userAuthToken    = "iamatoken"
	serviceAuthToken = "iamaservicetoken"
	collectionID     = "iamacollectionID"
	testHost         = "http://localhost:8080"
)

var (
	ctx          = context.Background()
	initialState = health.CreateCheckState(service)
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)

		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
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

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
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

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
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

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
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

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
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

func TestClient_GetArea(t *testing.T) {

	areaBody := `{
		  "code": "E92000001",
		  "date_end": null,
		  "date_start": "Thu, 01 Jan 2009 00: 00: 00 GMT",
		  "name": "England",
		  "name_welsh": "Lloegr",
		  "features": null,
		  "visible": true,
		 "area_type": "English"
		}`

	acceptedLang := "en-GB,en-US;q=0.9,en;q=0.8"
	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldNotBeNil)
	})

	Convey("When a area is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: areaBody})
		area, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldBeNil)
		So(area, ShouldResemble, AreaDetails{
			Code:        "E92000001",
			Name:        "England",
			DateStarted: "Thu, 01 Jan 2009 00: 00: 00 GMT",
			DateEnd:     "",
			WelshName:   "Lloegr",
			Visible:     true,
			AreaType:    "English",
		})
	})

	Convey("given a 200 status with valid empty body is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: "{}"})

		Convey("when GetArea is called", func() {
			instance, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)

			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, AreaDetails{})
			})
		})
	})
}

func TestClient_GetRelations(t *testing.T) {

	relationsBody := `[
			{
				"area_code": "E12000001",
				"area_name": "North East",
				"href": "/v1/area/E12000001"
			},
			{
				"area_code": "E12000002",
				"area_name": "North West",
				"href": "/v1/area/E12000002"
			},
			{
				"area_code": "E12000003",
				"area_name": "Yorkshire and The Humbe",
				"href": "/v1/area/E12000003"
			}
		]`
	expected := []Relation{Relation{AreaCode: "E12000001", AreaName: "North East", Href: "/v1/area/E12000001"}, Relation{AreaCode: "E12000002", AreaName: "North West", Href: "/v1/area/E12000002"}, Relation{AreaCode: "E12000003", AreaName: "Yorkshire and The Humbe", Href: "/v1/area/E12000003"}}
	acceptedLang := "en-GB,en-US;q=0.9,en;q=0.8"
	Convey("When a bad request is returned", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusBadRequest, Body: ""})
		_, err := mockedApi.GetRelations(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldNotBeNil)
	})

	Convey("When relations are returned", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: relationsBody})
		relations, err := mockedApi.GetRelations(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldBeNil)
		So(relations, ShouldResemble, expected)
	})
	Convey("given a 200 status with valid empty body is returned", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: "[]"})
		Convey("when GetRelations is called", func() {
			instance, err := mockedApi.GetRelations(ctx, userAuthToken, serviceAuthToken, collectionID, "92000001", acceptedLang)
			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, []Relation{})
			})
		})
	})
}

func TestClient_GetAncestors(t *testing.T) {
	expectedDidsburyEast := Ancestors{Ancestors: []Ancestor{Ancestor{Name: "Manchester", Level: "", Code: "E08000003", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}, Ancestor{Name: "North West", Level: "", Code: "E12000002", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}, Ancestor{Name: "England", Level: "", Code: "E92000001", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}}}
	expectedManchester := Ancestors{Ancestors: []Ancestor{Ancestor{Name: "North West", Level: "", Code: "E12000002", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}, Ancestor{Name: "England", Level: "", Code: "E92000001", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}}}
	expectedNorthWest := Ancestors{Ancestors: []Ancestor{Ancestor{Name: "England", Level: "", Code: "E92000001", Ancestors: []Ancestor{}, Siblings: []Ancestor{}, Children: []Ancestor{}}}}
	expectedEngland := Ancestors{Ancestors: []Ancestor{}}
	acceptedLang := "en-GB,en-US;q=0.9,en;q=0.8"

	testData := []string{
		`{ "name": "Didsbury East", "level": "", "code": "E05011362", "ancestors": [], "siblings": [], "children": [] }`,
		`{ "name": "Manchester", "level": "", "code": "E08000003", "ancestors": [], "siblings": [], "children": [] }`,
		`{ "name": "North West", "level": "", "code": "E12000002", "ancestors": [], "siblings": [], "children": [] }`,
		`{ "name": "England", "level": "", "code": "E92000001", "ancestors": [], "siblings": [], "children": [] }`,
	}

	Convey("Didsbury East code returns correct response body", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: getAncestry(testData[1], testData[2], testData[3])})
		ancestors, err := mockedApi.GetAncestors(ctx, userAuthToken, serviceAuthToken, collectionID, "E05011362", acceptedLang)
		So(err, ShouldBeNil)
		So(ancestors, ShouldResemble, expectedDidsburyEast)
	})
	Convey("Manchester code returns correct response body", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: getAncestry(testData[2], testData[3])})
		ancestors, err := mockedApi.GetAncestors(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldBeNil)
		So(ancestors, ShouldResemble, expectedManchester)
	})
	Convey("NorthWest code returns correct response body", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: getAncestry(testData[3])})
		ancestors, err := mockedApi.GetAncestors(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldBeNil)
		So(ancestors, ShouldResemble, expectedNorthWest)
	})
	Convey("England code returns correct response body", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: getAncestry("")})
		ancestors, err := mockedApi.GetAncestors(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001", acceptedLang)
		So(err, ShouldBeNil)
		So(ancestors, ShouldResemble, expectedEngland)
	})
	Convey("given a 200 status with valid empty body is returned", t, func() {
		mockedApi := getMockAreaAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: getAncestry("")})
		Convey("when GetRelations is called", func() {
			instance, err := mockedApi.GetAncestors(ctx, userAuthToken, serviceAuthToken, collectionID, "<RANDOM>", acceptedLang)
			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, Ancestors{Ancestors: []Ancestor{}})
			})
		})
	})
}

func getMockAreaAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
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

func newAreasClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	areasClient := NewWithHealthClient(healthClient)
	return areasClient
}

func getAncestry(ancestors ...string) string {
	var data []string
	for _, a := range ancestors {
		data = append(data, a)
	}
	fmt.Sprintf(`[%s]`, strings.Join(data, ","))
	return fmt.Sprintf(`[%s]`, strings.Join(data, ","))
}
