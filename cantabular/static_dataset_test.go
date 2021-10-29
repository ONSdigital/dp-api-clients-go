package cantabular_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

func TestStream(t *testing.T) {
	// responseBody := makeRequest(flag.Arg(0), flag.Args()[1:])

	Convey("RealTest", t, func() {
		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host:           "http://localhost:8491",
				ExtApiHost:     "http://localhost:8492",
				GraphQLTimeout: 10 * time.Second,
			},
			dphttp.NewClient(),
			nil,
		)

		req := cantabular.StaticDatasetQueryRequest{
			Dataset:   "Example",
			Variables: []string{"city", "siblings"},
		}
		responseBody, err := cantabularClient.StaticDatasetQuery(testCtx, req)
		So(err, ShouldBeNil)

		defer func() { _ = responseBody.Close() }()

		cantabular.GraphqlJSONToCSV(responseBody, os.Stdout)
		// So(true, ShouldBeFalse)
	})
}

// func TestStaticDatasetQueryHappy(t *testing.T) {

// 	Convey("Given a correct response from the /graphql endpoint", t, func() {
// 		testCtx := context.Background()

// 		mockHttpClient := &dphttp.ClienterMock{}
// 		mockGQLClient := &mock.GraphQLClientMock{
// 			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
// 				return nil
// 			},
// 		}

// 		cantabularClient := cantabular.NewClient(
// 			cantabular.Config{
// 				Host:       "cantabular.host",
// 				ExtApiHost: "cantabular.ext.host",
// 			},
// 			mockHttpClient,
// 			mockGQLClient,
// 		)

// 		Convey("When the StaticDatasetQuery method is called", func() {
// 			req := cantabular.StaticDatasetQueryRequest{}
// 			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

// 			Convey("No error should be returned", func() {
// 				So(err, ShouldBeNil)
// 			})
// 		})
// 	})
// }

// func TestStaticDatasetQueryUnHappy(t *testing.T) {

// 	Convey("Given the graphQL Client is not configured", t, func() {
// 		testCtx := context.Background()

// 		mockHttpClient := &dphttp.ClienterMock{}

// 		cantabularClient := cantabular.NewClient(
// 			cantabular.Config{
// 				Host: "cantabular.host",
// 			},
// 			mockHttpClient,
// 			nil,
// 		)

// 		Convey("When the StaticDatasetQuery method is called", func() {
// 			req := cantabular.StaticDatasetQueryRequest{}
// 			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
// 			So(err, ShouldNotBeNil)

// 			Convey("Status Code 503 Service Unavailable should be recoverable from error", func() {
// 				_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
// 				So(dperrors.StatusCode(err), ShouldEqual, http.StatusServiceUnavailable)
// 			})
// 		})
// 	})

// 	Convey("Given a GraphQL error from the /graphql endpoint", t, func() {
// 		testCtx := context.Background()

// 		mockHttpClient := &dphttp.ClienterMock{}
// 		mockGQLClient := &mock.GraphQLClientMock{
// 			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
// 				if q, ok := query.(*cantabular.StaticDatasetQuery); ok {
// 					q.Dataset.Table.Error = "I am error response"
// 					return nil
// 				}
// 				return errors.New("query could not be cast to correct type")
// 			},
// 		}

// 		cantabularClient := cantabular.NewClient(
// 			cantabular.Config{
// 				Host:       "cantabular.host",
// 				ExtApiHost: "cantabular.ext.host",
// 			},
// 			mockHttpClient,
// 			mockGQLClient,
// 		)

// 		Convey("When the StaticDatasetQuery method is called", func() {
// 			req := cantabular.StaticDatasetQueryRequest{}
// 			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

// 			Convey("An error should be returned with status code 400 Bad Request", func() {
// 				So(err, ShouldNotBeNil)
// 				So(dperrors.StatusCode(err), ShouldEqual, http.StatusBadRequest)
// 			})
// 		})
// 	})
// }
