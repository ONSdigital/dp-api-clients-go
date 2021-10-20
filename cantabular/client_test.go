package cantabular_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v3/cantabular"
	"github.com/ONSdigital/dp-healthcheck/v2/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChecker(t *testing.T) {
	testCtx := context.Background()

	Convey("Given a 200 OK response from the /v9/datasets endpoint", t, func() {

		mockHttpClient := createMockHttpClient(http.StatusOK)

		cantabularClient := cantabular.NewClient(
			cantabular.Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the GetCodebook method is called", func() {
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
	})

	Convey("Given a 500 response from the /v9/datasets endpoint", t, func() {

		mockHttpClient := createMockHttpClient(http.StatusInternalServerError)
		beforeCall := time.Now().UTC()

		cantabularClient := cantabular.NewClient(
			cantabular.Config{},
			&mockHttpClient,
			nil,
		)

		Convey("When the GetCodebook method is called", func() {
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
