package upload

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	userAuthToken    = "iamatoken"
	serviceAuthToken = "iamaservicetoken"
	collectionID     = "iamacollectionID"
	testHost         = "http://localhost:8080"
)

func TestClient_New(t *testing.T) {
	Convey("NewAPIClient creates a new API client with the expected URL and name", t, func() {
		uploadClient := NewAPIClient(testHost)
		So(uploadClient.URL(), ShouldEqual, testHost)
		So(uploadClient.HealthClient().Name, ShouldEqual, "upload-api")
	})

	Convey("Given an existing healthcheck client", t, func() {
		hcClient := health.NewClient("generic", testHost)
		Convey("The creating a new upload API client providing it, results in a new client with the expected URL and name", func() {
			uploadClient := NewWithHealthClient(hcClient)
			So(uploadClient.URL(), ShouldEqual, testHost)
			So(uploadClient.HealthClient().Name, ShouldEqual, "upload-api")
		})
	})
}
