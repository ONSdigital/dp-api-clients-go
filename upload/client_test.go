package upload_test

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	Convey("Given the upload service is health", t, func() {
		timePriorHealthCheck := time.Now()
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		state := health.CreateCheckState("testing")

		Convey("When we check that state of the service", func() {
			c := upload.NewAPIClient(s.URL)
			c.Checker(context.Background(), &state)

			So(state.Status(), ShouldEqual, healthcheck.StatusOK)
			So(state.StatusCode(), ShouldEqual, 200)
			So(state.Message(), ShouldContainSubstring, "is ok")

			So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
			So(*state.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
			So(state.LastFailure(), ShouldBeNil)
		})
	})

	Convey("Given the upload service is failing", t, func() {
		timePriorHealthCheck := time.Now()
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
		defer s.Close()

		state := health.CreateCheckState("testing")

		Convey("When we check the state of the service", func() {
			c := upload.NewAPIClient(s.URL)
			c.Checker(context.Background(), &state)

			So(state.Status(), ShouldEqual, healthcheck.StatusCritical)
			So(state.StatusCode(), ShouldEqual, 500)
			So(state.Message(), ShouldContainSubstring, "unavailable or non-functioning")

			So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
			So(state.LastSuccess(), ShouldBeNil)
			So(*state.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
		})
	})
}
