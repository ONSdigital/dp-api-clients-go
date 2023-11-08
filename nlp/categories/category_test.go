package category

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/categories/models"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	. "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:28800"

var (
	initialTestState = healthcheck.CreateCheckState(service)

	categoryResults = []models.Category{
		{
			Code: []string{
				"test1",
			},
			Score: 12.123,
		},
		{
			Code: []string{
				"test2",
			},
			Score: 12.123,
		},
	}
)

func TestHealthCheckerClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	timePriorHealthCheck := time.Now().UTC()
	path := "/health"

	Convey("Given clienter.Do returns an error", t, func() {
		clientError := errors.New("unexpected error")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		categoryAPIClient := newCategoryAPIClient(t, httpClient)
		check := initialTestState

		Convey("When category API client Checker is called", func() {
			err := categoryAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("Given a 500 response for health check", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		categoryAPIClient := newCategoryAPIClient(t, httpClient)
		check := initialTestState

		Convey("When category API client Checker is called", func() {
			err := categoryAPIClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("Then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, health.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+healthcheck.StatusMessage[health.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("And client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestGetCategory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given request to find category results", t, func() {
		body, err := json.Marshal(categoryResults)
		if err != nil {
			t.Errorf("failed to setup test data, error: %v", err)
		}

		httpClient := newMockHTTPClient(
			&http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewReader(body)),
			},
			nil)

		categoryAPIClient := newCategoryAPIClient(t, httpClient)

		Convey("When GetCategory is called", func() {
			query := url.Values{}
			query.Add("query", "census")
			resp, err := categoryAPIClient.GetCategory(ctx, Options{Query: query})

			Convey("Then the expected response body is returned", func() {
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)

				Convey("And no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And client.Do should be called once with the expected parameters", func() {
						doCalls := httpClient.DoCalls()
						So(doCalls, ShouldHaveLength, 1)
						So(doCalls[0].Req.Method, ShouldEqual, "GET")
						So(doCalls[0].Req.URL.Path, ShouldEqual, "/categories")
						So(doCalls[0].Req.URL.Query().Get("query"), ShouldEqual, "census")
						So(doCalls[0].Req.Header["Authorization"], ShouldBeEmpty)
					})
				})
			})
		})
	})
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newCategoryAPIClient(t *testing.T, httpClient *dphttp.ClienterMock) *Client {
	healthClient := healthcheck.NewClientWithClienter(service, testHost, httpClient)
	return NewWithHealthClient(healthClient)
}
