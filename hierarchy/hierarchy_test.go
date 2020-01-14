package hierarchy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

var ctx = context.Background()

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
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

func TestClient_GetHealth(t *testing.T) {

	Convey("Given a healthy hierarchy api is running", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Healthcheck is called", func() {
			serviceName, err := cli.Healthcheck()

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
				So(serviceName, ShouldEqual, service)
				So(len(mockRCHTTPCli.GetCalls()), ShouldEqual, 1)
			})
		})
	})

	Convey("Given hierarchy api does not contain a healthcheck endpoint", t, func() {
		mockErr := errors.New("endpoint not found")
		mockRCHTTPCli := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, mockErr
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Healthcheck is called", func() {
			serviceName, err := cli.Healthcheck()

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
				So(serviceName, ShouldEqual, service)
				So(len(mockRCHTTPCli.GetCalls()), ShouldEqual, 2)
			})
		})
	})

	Convey("Given hierarchy api is not running", t, func() {
		mockErr := errors.New("internal server error")
		mockRCHTTPCli := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, mockErr
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Healthcheck is called", func() {
			serviceName, err := cli.Healthcheck()

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, mockErr.Error())
				So(serviceName, ShouldEqual, service)
				So(len(mockRCHTTPCli.GetCalls()), ShouldEqual, 1)
			})
		})
	})
}
