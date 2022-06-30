package cantabular

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	. "github.com/smartystreets/goconvey/convey"
)

type testError struct {
	err        error
	statusCode int
}

func (e *testError) Error() string {
	if e.err == nil {
		return "nil"
	}
	return e.err.Error()
}

func (e *testError) Unwrap() error {
	return e.err
}

func (e *testError) Code() int {
	return e.statusCode
}

func TestChecker(t *testing.T) {
	testCtx := context.Background()

	Convey("Given that a 200 OK response is returned", t, func() {

		mockHttpClient := createMockHttpClient(http.StatusOK)

		cantabularClient := NewClient(
			Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, fmt.Sprintf("/%s/datasets", SoftwareVersion))
			})

			Convey("Then the CheckState is updated to the expected OK state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.Message(), ShouldEqual, "cantabular is ok")
				So(check.LastFailure(), ShouldBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected OK state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.Message(), ShouldEqual, "cantabularAPIExt is ok")
				So(check.LastFailure(), ShouldBeNil)
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given that a 500 response is returned", t, func() {

		mockHttpClient := createMockHttpClient(http.StatusInternalServerError)
		beforeCall := time.Now().UTC()

		cantabularClient := NewClient(
			Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, fmt.Sprintf("/%s/datasets", SoftwareVersion))
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "cantabular functionality is unavailable or non-functioning")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "cantabularAPIExt functionality is unavailable or non-functioning")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an http client that returns an error", t, func() {

		mockHttpClient := dphttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return nil, errors.New("mock error")
			},
		}
		beforeCall := time.Now().UTC()

		cantabularClient := NewClient(
			Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, fmt.Sprintf("/%s/datasets", SoftwareVersion))
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "failed to make request: mock error")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, Service)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "failed to make request: mock error")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestStatusCode(t *testing.T) {
	client := NewClient(
		Config{},
		nil,
		nil,
	)

	Convey("Given an error with embedded status code", t, func() {
		err := &testError{
			statusCode: http.StatusTeapot,
		}

		Convey("When StatusCode(err) is called", func() {
			status := client.StatusCode(err)
			expected := http.StatusTeapot

			So(status, ShouldEqual, expected)
		})
	})
}

func createMockHttpClient(statusCode int) dphttp.ClienterMock {
	return dphttp.ClienterMock{
		GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
			return Response(
				nil,
				statusCode,
			), nil
		},
	}
}
