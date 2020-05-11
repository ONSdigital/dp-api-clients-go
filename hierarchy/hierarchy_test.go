package hierarchy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx          = context.Background()
	testHost     = "http://localhost:8080"
	initialState = health.CreateCheckState(service)
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func TestClient_HealthChecker(t *testing.T) {
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hierarchyClient := New(testHost)
		hierarchyClient.cli = clienter
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hierarchyClient := New(testHost)
		hierarchyClient.cli = clienter
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hierarchyClient := New(testHost)
		hierarchyClient.cli = clienter
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 429,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hierarchyClient := New(testHost)
		hierarchyClient.cli = clienter
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hierarchyClient := New(testHost)
		hierarchyClient.cli = clienter
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func getMockHierarchyAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) (*httptest.Server, *Client) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	return ts, New(ts.URL)
}

func TestClient_GetRoot(t *testing.T) {
	instanceID := "foo"
	name := "bar"
	model := `{"label":"","has_data":true}`
	Convey("When bad request is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldNotBeNil)
		ts.Close()
	})

	Convey("When server error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.cli.SetMaxRetries(2)
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldNotBeNil)
		ts.Close()
	})
	Convey("When a hierarchy-instance is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldBeNil)
		ts.Close()
	})

}

func TestClient_GetChild(t *testing.T) {
	instanceID := "foo"
	name := "bar"
	code := "baz"
	model := `{"label":"","has_data":true}`
	Convey("When bad request is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldNotBeNil)
		ts.Close()
	})

	Convey("When server error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.cli.SetMaxRetries(2)
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldNotBeNil)
		ts.Close()
	})

	Convey("When a hierarchy-instance is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldBeNil)
		ts.Close()
	})
}
func TestClient_GetHierarchy(t *testing.T) {
	path := "/hierarchies/foo/bar"
	model := `{"label":"","has_data":true}`

	Convey("When bad request is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.getHierarchy(ctx, path)
		So(err, ShouldNotBeNil)
		ts.Close()
	})

	Convey("When server error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "qux"})
		mockedAPI.cli.SetMaxRetries(2)
		_, err := mockedAPI.getHierarchy(ctx, path)
		So(err, ShouldNotBeNil)
		ts.Close()
	})

	Convey("When a hierarchy-instance is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: model})
		_, err := mockedAPI.getHierarchy(ctx, path)
		So(err, ShouldBeNil)
		ts.Close()
	})

}
