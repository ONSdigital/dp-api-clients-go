package search

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testHost         = "http://localhost:8080"
)

var (
	ctx          = context.Background()
	initialState = health.CreateCheckState(service)
)

var checkResponseBase = func(mockdphttpCli *dphttp.ClienterMock, expectedMethod string, expectedUri string) {
	So(len(mockdphttpCli.DoCalls()), ShouldEqual, 1)
	So(mockdphttpCli.DoCalls()[0].Req.URL.RequestURI(), ShouldEqual, expectedUri)
	So(mockdphttpCli.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
}

func createHTTPClientMock(retCode int, body []byte) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: retCode,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
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

		searchClient := NewClient(testHost)
		searchClient.cli = clienter
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
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

	Convey("given clienter.Do returns 400 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 400,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		searchClient := NewClient(testHost)
		searchClient.cli = clienter
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 400)
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

		searchClient := NewClient(testHost)
		searchClient.cli = clienter
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
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
}

func TestClient_GetSearch(t *testing.T) {
	Convey("given a 200 status is returned with an empty result list", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/empty_results.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := Client{cli: mockdphttpCli, url: "http://localhost:8080"}

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("q", "a")
			r, err := cli.GetSearch(ctx, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 0)
				So(r.ContentTypes, ShouldBeEmpty)
				So(r.Items, ShouldBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/search?q=a")
			})
		})
	})

	Convey("given a 200 status is returned with list of search results", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/results.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := Client{cli: mockdphttpCli, url: "http://localhost:8080"}

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("q", "housing")
			r, err := cli.GetSearch(ctx, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 5)
				So(r.Items, ShouldNotBeEmpty)
				So(r.ContentTypes, ShouldNotBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/search?q=housing")
			})
		})
	})

	Convey("given a 400 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusBadRequest, nil)
		cli := Client{cli: mockdphttpCli, url: "http://localhost:8080"}

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "a")
			_, err := cli.GetSearch(ctx, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response from dp-search-query - should be: 200, got: 400, path: "+testHost+"/search?limit=a").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/search?limit=a")
			})
		})
	})

	Convey("given a 500 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusInternalServerError, nil)
		cli := Client{cli: mockdphttpCli, url: "http://localhost:8080"}

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "housing")
			_, err := cli.GetSearch(ctx, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response from dp-search-query - should be: 200, got: 500, path: "+testHost+"/search?limit=housing").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/search?limit=housing")
			})
		})
	})
}

