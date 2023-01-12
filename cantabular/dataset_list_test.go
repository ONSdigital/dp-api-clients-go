package cantabular_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListDatasetsHappy(t *testing.T) {
	Convey("Given a valid response from the /graphql endpoint", t, func() {
		ctx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyListDatasets, http.StatusOK)

		Convey("When ListDatasets is called", func() {
			resp, err := cantabularClient.ListDatasets(ctx)
			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryListDatasets,
					cantabular.QueryData{},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedListDatasets)
			})
		})

	})
}

func TestListDatasetsUnhappy(t *testing.T) {
	ctx := context.Background()

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		_, client := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetAreas is called", func() {
			resp, err := client.ListDatasets(ctx)

			Convey("Then the expected error is returned", func() {
			})
			So(client.StatusCode(err), ShouldResemble, http.StatusInternalServerError)

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

// mockRespBodyListDatasets is a successful 'list datasets' response
var mockRespBodyListDatasets = `
{
	"data": {
		"datasets": [
			{
				"name": "dataset_1",
				"label": "dataset 1"
			},
			{
				"name": "dataset_2",
				"label": "dataset 2"
			}
		]
	}
}
`
var expectedListDatasets = cantabular.ListDatasetsResponse{
	Datasets: []gql.Dataset{
		{
			Name:  "dataset_1",
			Label: "dataset 1",
		},
		{
			Name:  "dataset_2",
			Label: "dataset 2",
		},
	},
}
