package cantabular_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v3/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v3/cantabular/mock"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v3/errors"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStaticDatasetQueryHappy(t *testing.T) {

	Convey("Given a correct response from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}
		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				return nil
			},
		}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host:       "cantabular.host",
				ExtApiHost: "cantabular.ext.host",
			},
			mockHttpClient,
			mockGQLClient,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestStaticDatasetQueryUnHappy(t *testing.T) {

	Convey("Given the graphQL Client is not configured", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host: "cantabular.host",
			},
			mockHttpClient,
			nil,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
			So(err, ShouldNotBeNil)

			Convey("Status Code 503 Service Unavailable should be recoverable from error", func() {
				_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
				So(dperrors.StatusCode(err), ShouldEqual, http.StatusServiceUnavailable)
			})
		})
	})

	Convey("Given a GraphQL error from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}
		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				if q, ok := query.(*cantabular.StaticDatasetQuery); ok {
					q.Dataset.Table.Error = "I am error response"
					return nil
				}
				return errors.New("query could not be cast to correct type")
			},
		}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host:       "cantabular.host",
				ExtApiHost: "cantabular.ext.host",
			},
			mockHttpClient,
			mockGQLClient,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

			Convey("An error should be returned with status code 400 Bad Request", func() {
				So(err, ShouldNotBeNil)
				So(dperrors.StatusCode(err), ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
