package cantabular_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/mock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListDatasetsHappy(t *testing.T) {

	Convey("Should request dataset names from cantabular", t, func() {

		fakeConfig := cantabular.Config{
			Host:       "cantabular.host",
			ExtApiHost: "cantabular.ext.host",
		}

		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				DatasetListQuery := query.(*cantabular.ListDatasetsQuery)
				DatasetListQuery.Datasets = []cantabular.ListDatasetsItem{
					{Name: "dataset 1"},
					{Name: "dataset 2"},
				}
				return nil
			},
		}

		cantabularClient := cantabular.NewClient(fakeConfig, nil, mockGQLClient)
		list, err := cantabularClient.ListDatasets(context.Background())

		actualQueryCall := mockGQLClient.QueryCalls()[0]
		SoMsg("context should be passed through", actualQueryCall.Ctx, ShouldEqual, context.Background())
		SoMsg("no error should be returned", err, ShouldBeNil)
		expectedNames := []string{"dataset 1", "dataset 2"}
		SoMsg("returned list of names should match expected", list, ShouldResemble, expectedNames)
	})
}

func TestListDatasetsUnhappy(t *testing.T) {

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
			actualList, actualErr := cantabularClient.ListDatasets(context.Background())
			SoMsg("error should be populated", actualErr, ShouldEqual, expectedError)
			SoMsg("DatasetList returned should be nil", actualList, ShouldBeNil)
		})
	})
}
