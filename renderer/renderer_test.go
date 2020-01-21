package renderer

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
)

var ctx = context.Background()

const (
	testHost = "http://localhost:8080"
)

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		renderer := New(testHost)
		renderer.cli = clienter

		Convey("when renderer.Checker is called", func() {
			check, err := renderer.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 0)
				So(check.Message, ShouldEqual, clientError.Error())
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		clienter := &rchttp.ClienterMock{
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

		renderer := New(testHost)
		renderer.cli = clienter

		Convey("when renderer.Checker is called", func() {
			check, err := renderer.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 500)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		clienter := &rchttp.ClienterMock{
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

		renderer := New(testHost)
		renderer.cli = clienter

		Convey("when renderer.Checker is called", func() {
			check, err := renderer.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 404)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
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
		clienter := &rchttp.ClienterMock{
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

		renderer := New(testHost)
		renderer.cli = clienter

		Convey("when renderer.Checker is called", func() {
			check, err := renderer.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode, ShouldEqual, 429)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		clienter := &rchttp.ClienterMock{
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

		renderer := New(testHost)
		renderer.cli = clienter

		Convey("when renderer.Checker is called", func() {
			check, err := renderer.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode, ShouldEqual, 200)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}
