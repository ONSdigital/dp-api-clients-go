package hierarchy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
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

func TestErrInvalidHierarchyAPIResponse(t *testing.T) {

	Convey("Given an error created with NewErrInvalidHierarchyAPIResponse", t, func() {
		err := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusBadRequest, "/hierarchies/foo/bar")

		Convey("Then the error is an ErrInvalidHierarchyAPIResponse with the expected fields", func() {
			So(err, ShouldResemble, &ErrInvalidHierarchyAPIResponse{
				expectedCode: http.StatusOK,
				actualCode:   http.StatusBadRequest,
				uri:          "/hierarchies/foo/bar",
			})
		})

		Convey("Then Errors() returns the expected error message", func() {
			So(err.Error(), ShouldResemble, "invalid response from hierarchy api - should be: 200, got: 400, path: /hierarchies/foo/bar")
		})

		Convey("Then Code() returns the actual http stauts code", func() {
			So(err.(*ErrInvalidHierarchyAPIResponse).Code(), ShouldEqual, http.StatusBadRequest)
		})
	})
}

func TestClient_HealthChecker(t *testing.T) {
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given the http client returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		hierarchyClient := newHierarchyClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 500 response", t, func() {
		clienter := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)

		hierarchyClient := newHierarchyClient(clienter)
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, http.StatusInternalServerError)
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

	Convey("given a 404 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusNotFound}, nil)
		hierarchyClient := newHierarchyClient(httpClient)
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, http.StatusNotFound)
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

	Convey("given a 429 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusTooManyRequests}, nil)
		hierarchyClient := newHierarchyClient(httpClient)
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, http.StatusTooManyRequests)
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

	Convey("given a 200 OK response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusOK}, nil)
		hierarchyClient := newHierarchyClient(httpClient)
		check := initialState

		Convey("when hierarchyClient.Checker is called", func() {
			err := hierarchyClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, http.StatusOK)
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

	Convey("Given a bad request API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusBadRequest, Body: ""})
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusBadRequest, "/hierarchies/foo/bar")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a server error API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusInternalServerError, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetRoot(ctx, instanceID, name)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusInternalServerError, "/hierarchies/foo/bar")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a 200 OK response, then the expected hierarchy-instance is returned with no error", t, func() {
		model := `{"label":"testModel","has_data":true,"order":0}`
		expectedOrder := 0
		expectedModel := Model{
			Label:   "testModel",
			HasData: true,
			Order:   &expectedOrder,
		}

		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: model})
		m, err := mockedAPI.GetRoot(ctx, instanceID, name)
		So(err, ShouldBeNil)
		So(m, ShouldResemble, expectedModel)
		ts.Close()
	})
}

func TestClient_GetChild(t *testing.T) {
	instanceID := "foo"
	name := "bar"
	code := "baz"

	Convey("Given a bad request API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusBadRequest, Body: ""})
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusBadRequest, "/hierarchies/foo/bar/baz")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a server error API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusInternalServerError, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusInternalServerError, "/hierarchies/foo/bar/baz")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a hierarchy-instance API response, then the expected model is returned with no error", t, func() {
		model := `{"label":"testChild","has_data":true,"order":321}`
		expectedOrder := 321
		expectedModel := Model{
			Label:   "testChild",
			HasData: true,
			Order:   &expectedOrder,
		}

		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: model})
		m, err := mockedAPI.GetChild(ctx, instanceID, name, code)
		So(err, ShouldBeNil)
		So(m, ShouldResemble, expectedModel)
		ts.Close()
	})
}

func TestClient_GetHierarchy(t *testing.T) {
	path := "/hierarchies/foo/bar"

	Convey("Given a bad request API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusBadRequest, Body: ""})
		_, err := mockedAPI.getHierarchy(ctx, path)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusBadRequest, "/hierarchies/foo/bar")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a server error API response, then the expected error is returned", t, func() {
		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusInternalServerError, Body: "qux"})
		mockedAPI.hcCli.Client.SetMaxRetries(2)
		_, err := mockedAPI.getHierarchy(ctx, path)
		expectedErr := NewErrInvalidHierarchyAPIResponse(http.StatusOK, http.StatusInternalServerError, "/hierarchies/foo/bar")
		So(err, ShouldResemble, expectedErr)
		ts.Close()
	})

	Convey("Given a hierarchy-instance API response, then the expected model is returned with no error", t, func() {
		model := `{"label":"testHierarchy","has_data":true,"order":123}`
		expectedOrder := 123
		expectedModel := Model{
			Label:   "testHierarchy",
			HasData: true,
			Order:   &expectedOrder,
		}

		ts, mockedAPI := getMockHierarchyAPI(http.Request{Method: http.MethodGet}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: model})
		m, err := mockedAPI.getHierarchy(ctx, path)
		So(err, ShouldBeNil)
		So(m, ShouldResemble, expectedModel)
		ts.Close()
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newHierarchyClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, httpClient)
	hierarchyClient := NewWithHealthClient(healthClient)
	return hierarchyClient
}
