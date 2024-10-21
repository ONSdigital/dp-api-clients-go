package tablerenderer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/http"
)

const (
	testHost = "http://localhost:8080"
)

func TestClientNew(t *testing.T) {
	Convey("New creates a new client with the expected URL and name", t, func() {
		client := New(testHost)
		So(client.URL(), ShouldEqual, testHost)
		So(client.HealthClient().Name, ShouldEqual, "table-renderer")
	})

	Convey("Given an existing healthcheck client", t, func() {
		hcClient := health.NewClient("generic", testHost)
		Convey("When creating a new table rednerer client providing it", func() {
			client := NewWithHealthClient(hcClient)
			Convey("Then it returns a new client with the expected URL and name", func() {
				So(client.URL(), ShouldEqual, testHost)
				So(client.HealthClient().Name, ShouldEqual, "table-renderer")
			})
		})
	})
}

func TestRender(t *testing.T) {
	json := []byte("{ }")
	html := []byte("<html/>")

	Convey("Given that 200 OK is returned by the service", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(html)),
		}, nil)
		client := newTableRendererClient(httpClient)

		Convey("When Render is called", func() {
			response, err := client.Render(context.Background(), "html", json)

			Convey("Then the expected call to the table-renderer is made", func() {
				expectedUrl := fmt.Sprintf("%s/render/html", testHost)
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodPost)
			})
			Convey("And the expected result is returned without error", func() {
				So(err, ShouldBeNil)
				So(response, ShouldResemble, html)
			})
		})
	})

	Convey("Given that 404 is returned by the service ", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("URL not found"))),
		}, nil)
		client := newTableRendererClient(httpClient)
		Convey("When Render is called", func() {
			response, err := client.Render(context.Background(), "csv", json)
			Convey("Then the expected call to the table-renderer is made", func() {
				expectedUrl := fmt.Sprintf("%s/render/csv", testHost)
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodPost)
			})
			Convey("And an error is returned", func() {
				So(err, ShouldResemble, ErrInvalidTableRendererResponse{responseCode: 404})
				So(response, ShouldBeNil)
			})
		})
	})

	Convey("Given an http client that fails to perform a request", t, func() {
		errorString := "table renderer error"
		httpClient := newMockHTTPClient(nil, errors.New(errorString))
		client := newTableRendererClient(httpClient)

		Convey("When Render is called", func() {
			response, err := client.Render(context.Background(), "xlsx", json)
			Convey("Then the expected call to the table-renderer is made", func() {
				expectedUrl := fmt.Sprintf("%s/render/xlsx", testHost)
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodPost)
			})
			Convey("And an error is returned", func() {
				So(err.Error(), ShouldResemble, errorString)
				So(response, ShouldBeNil)
			})
		})
	})
}

func newTableRendererClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	return NewWithHealthClient(healthClient)
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
	}
}
