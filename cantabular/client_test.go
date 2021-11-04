package cantabular_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChecker(t *testing.T) {
	testCtx := context.Background()

	Convey("Given that a 200 OK response is returned", t, func() {

		mockHttpClient := createMockHttpClient(http.StatusOK)

		cantabularClient := cantabular.NewClient(
			cantabular.Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/v9/datasets")
			})

			Convey("Then the CheckState is updated to the expected OK state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.Message(), ShouldEqual, "cantabular is ok")
				So(check.LastFailure(), ShouldBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected OK state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
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

		cantabularClient := cantabular.NewClient(
			cantabular.Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/v9/datasets")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "cantabular functionality is unavailable or non-functioning")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
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

		cantabularClient := cantabular.NewClient(
			cantabular.Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the Checker method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.Checker(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/v9/datasets")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "failed to make request: mock error")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
		})

		Convey("When the CheckerApiExt method is called", func() {
			check := healthcheck.NewCheckState(cantabular.Service)
			err := cantabularClient.CheckerAPIExt(testCtx, check)
			So(err, ShouldBeNil)

			Convey("Then the expected endpoint is called", func() {
				So(mockHttpClient.GetCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.GetCalls()[0].URL, ShouldEqual, "/graphql?query={}")
			})

			Convey("Then the CheckState is updated to the expected CRITICAL state", func() {
				So(check.Name(), ShouldEqual, cantabular.Service)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.Message(), ShouldEqual, "failed to make request: mock error")
				So(*check.LastFailure(), ShouldHappenBetween, beforeCall, time.Now().UTC())
				So(err, ShouldBeNil)
			})
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
