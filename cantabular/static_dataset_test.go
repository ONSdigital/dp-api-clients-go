package cantabular_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

var testCsv = `count,City,Number of siblings
1,London,No siblings
0,London,1 sibling
0,London,2 siblings
1,London,3 siblings
0,London,4 siblings
0,London,5 siblings
0,London,6 or more siblings
0,Liverpool,No siblings
0,Liverpool,1 sibling
0,Liverpool,2 siblings
0,Liverpool,3 siblings
1,Liverpool,4 siblings
0,Liverpool,5 siblings
0,Liverpool,6 or more siblings
0,Belfast,No siblings
0,Belfast,1 sibling
1,Belfast,2 siblings
0,Belfast,3 siblings
0,Belfast,4 siblings
1,Belfast,5 siblings
1,Belfast,6 or more siblings
`

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

		out := ""
		consume := func(ctx context.Context, r io.Reader) error {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				out += fmt.Sprintln(line)
			}
			return nil
		}

		req := cantabular.StaticDatasetQueryRequest{
			Dataset:   "Example",
			Variables: []string{"city", "siblings"},
		}
		err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
		So(err, ShouldBeNil)

		So(out, ShouldResemble, testCsv)
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
