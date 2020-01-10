package health

import (
	"context"
	"testing"
	"time"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHealth_GetCheck(t *testing.T) {
	defaultTime := time.Now().UTC()
	ctx := context.Background()

	Convey("Given the healthcheck status is OK return an OK check object", t, func() {
		check := getCheck(ctx, "api", health.StatusOK, "", 200)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 200)
		So(check.Status, ShouldEqual, health.StatusOK)
		So(check.Message, ShouldEqual, healthyMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldHappenAfter, defaultTime)
		So(check.LastFailure, ShouldEqual, unixTime)
	})

	Convey("Given the healthcheck status is WARNING return a WARNING check object", t, func() {
		check := getCheck(ctx, "api", health.StatusWarning, "", 200)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 200)
		So(check.Status, ShouldEqual, health.StatusWarning)
		So(check.Message, ShouldEqual, warningMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

	Convey("Given the healthcheck status is CRITICAL return a CRITICAL check object", t, func() {
		check := getCheck(ctx, "api", health.StatusCritical, "", 200)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 200)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, criticalMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

	Convey("Given the healthcheck endpoint does not exist return a CRITICAL check object", t, func() {
		errorMessage := "error: not found"
		check := getCheck(ctx, "api", health.StatusCritical, errorMessage, 404)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 404)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, notFoundMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

	Convey("Given request could not connect to api return a CRITICAL check object", t, func() {
		code := 500
		errorMessage := "error: internal service error"
		check := getCheck(ctx, "api", health.StatusCritical, errorMessage, code)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, code)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, errorMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

	Convey("Given request could not connect due to too many requests to the api "+
		"(denial of service), return a CRITICAL check object", t, func() {
		code := 429
		errorMessage := "error: too many requests"
		check := getCheck(ctx, "api", health.StatusCritical, errorMessage, code)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, code)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, errorMessage)
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})
}
