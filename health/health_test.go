package health

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

const apiName = "api"

var ctx = context.Background()

func getMockAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))

	api := NewClient(apiName, ts.URL)

	return api
}

func TestClient_GetOutput(t *testing.T) {
	defaultTime := time.Now().UTC()

	Convey("When health endpoint returns status OK", t, func() {
		mockedAPI := getMockAPI(
			http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 200, Body: "{\"status\": \"OK\"}"},
		)

		check, err := mockedAPI.Checker(ctx)
		So(check.Name, ShouldEqual, apiName)
		So(check.StatusCode, ShouldEqual, 200)
		So(check.Status, ShouldEqual, health.StatusOK)
		So(check.Message, ShouldEqual, statusDescription[health.StatusOK])
		So(*check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastFailure, ShouldBeNil)
		So(*check.LastSuccess, ShouldHappenAfter, defaultTime)
		So(err, ShouldBeNil)
	})

	Convey("When health endpoint returns status Warning", t, func() {
		mockedAPI := getMockAPI(
			http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 429, Body: "{\"status\": \"WARNING\"}"},
		)

		check, err := mockedAPI.Checker(ctx)
		So(check.Name, ShouldEqual, apiName)
		So(check.StatusCode, ShouldEqual, 429)
		So(check.Status, ShouldEqual, health.StatusWarning)
		So(check.Message, ShouldEqual, statusDescription[health.StatusWarning])
		So(*check.LastChecked, ShouldHappenAfter, defaultTime)
		So(*check.LastFailure, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("When health endpoint returns status Critical", t, func() {
		mockedAPI := getMockAPI(
			http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: "{\"status\": \"CRITICAL\"}"},
		)

		check, err := mockedAPI.Checker(ctx)
		So(check.Name, ShouldEqual, apiName)
		So(check.StatusCode, ShouldEqual, 500)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, statusDescription[health.StatusCritical])
		So(*check.LastChecked, ShouldHappenAfter, defaultTime)
		So(*check.LastFailure, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("When health endpoint is not implemented a status code of 404 is returned", t, func() {
		mockedAPI := getMockAPI(
			http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 404, Body: ""},
		)

		check, err := mockedAPI.Checker(ctx)
		So(check.Name, ShouldEqual, apiName)
		So(check.StatusCode, ShouldEqual, 404)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, statusDescription[health.StatusCritical])
		So(*check.LastChecked, ShouldHappenAfter, defaultTime)
		So(*check.LastFailure, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldBeNil)
		So(err, ShouldBeNil)
	})

	Convey("When an api is unavailable a status code of 500 is returned", t, func() {
		mockedAPI := getMockAPI(
			http.Request{Method: "GET"},
			MockedHTTPResponse{StatusCode: 500, Body: ""},
		)

		check, err := mockedAPI.Checker(ctx)
		So(check.Name, ShouldEqual, apiName)
		So(check.StatusCode, ShouldEqual, 500)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, statusDescription[health.StatusCritical])
		So(*check.LastChecked, ShouldHappenAfter, defaultTime)
		So(*check.LastFailure, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldBeNil)
		So(err, ShouldBeNil)
	})
}
