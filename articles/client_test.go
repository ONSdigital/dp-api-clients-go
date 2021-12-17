package articles

import (
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClientNew(t *testing.T) {
	Convey("NewAPIClient creates a new API client with the expected URL and name", t, func() {
		client := NewAPIClient(testHost)
		So(client.URL(), ShouldEqual, testHost)
		So(client.HealthClient().Name, ShouldEqual, "articles-api")
	})

	Convey("Given an existing healthcheck client", t, func() {
		hcClient := health.NewClient("generic", testHost)
		Convey("When creating a new article API client providing it", func() {
			client := NewWithHealthClient(hcClient)
			Convey("Then it returns a new client with the expected URL and name", func() {
				So(client.URL(), ShouldEqual, testHost)
				So(client.HealthClient().Name, ShouldEqual, "articles-api")
			})
		})
	})
}
