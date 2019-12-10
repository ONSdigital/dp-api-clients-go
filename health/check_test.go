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

	Convey("Given a 200 status code return OK health check object", t, func() {
		check := getCheck(&ctx, "api", 200)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 200)
		So(check.Status, ShouldEqual, health.StatusOK)
		So(check.Message, ShouldEqual, statusDescription[health.StatusOK])
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldHappenAfter, defaultTime)
		So(check.LastFailure, ShouldEqual, unixTime)
	})

	Convey("Given a 429 status code return Warning health check object", t, func() {
		check := getCheck(&ctx, "api", 429)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 429)
		So(check.Status, ShouldEqual, health.StatusWarning)
		So(check.Message, ShouldEqual, statusDescription[health.StatusWarning])
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

	Convey("Given a 404 status code return Critical health check object", t, func() {
		check := getCheck(&ctx, "api", 404)

		So(check.Name, ShouldEqual, "api")
		So(check.StatusCode, ShouldEqual, 404)
		So(check.Status, ShouldEqual, health.StatusCritical)
		So(check.Message, ShouldEqual, statusDescription[health.StatusCritical])
		So(check.LastChecked, ShouldHappenAfter, defaultTime)
		So(check.LastSuccess, ShouldEqual, unixTime)
		So(check.LastFailure, ShouldHappenAfter, defaultTime)
	})

}
