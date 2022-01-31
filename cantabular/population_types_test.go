package cantabular_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlobsHappy(t *testing.T) {

	Convey("Population types should request dataset names from cantabular", t, func() {

		fakeConfig := cantabular.Config{
			Host:       "cantabular.host",
			ExtApiHost: "cantabular.ext.host",
		}

		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				BlobsQuery := query.(*cantabular.BlobQuery)
				BlobsQuery.Datasets = []cantabular.BlobQueryDataset{
					{Name: "blob 1"},
					{Name: "blob 2"},
				}
				return nil
			},
		}

		cantabularClient := cantabular.NewClient(fakeConfig, nil, mockGQLClient)
		Blobs, err := cantabularClient.GetBlobs(context.Background())

		actualQueryCall := mockGQLClient.QueryCalls()[0]
		SoMsg("context should be passed through", actualQueryCall.Ctx, ShouldEqual, context.Background())
		SoMsg("no error should be returned", err, ShouldBeNil)
		expectedNames := []string{"blob 1", "blob 2"}
		SoMsg("returned list of names should match expected", Blobs, ShouldResemble, expectedNames)
	})
}

func TestBlobsUnhappy(t *testing.T) {

	fakeConfig := cantabular.Config{
		Host:       "cantabular.host",
		ExtApiHost: "cantabular.ext.host",
	}

	Convey("Given cantabular returns an error", t, func() {

		expectedError := errors.New("nope")
		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				return expectedError
			},
		}
		cantabularClient := cantabular.NewClient(fakeConfig, nil, mockGQLClient)

		Convey("Population types should return an error", func() {
			actualBlobs, actualErr := cantabularClient.GetBlobs(context.Background())
			SoMsg("error should be populated", actualErr, ShouldEqual, expectedError)
			SoMsg("Blobs returned should be nil", actualBlobs, ShouldBeNil)
		})
	})
}
