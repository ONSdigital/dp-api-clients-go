package cantabularmetadata_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabularmetadata"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabularmetadata/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDefaultClassificationHappy(t *testing.T) {
	Convey("Given a correct GetDefaultClassication response from the /graphql endpoint", t, func() {
		ctx := context.Background()
		httpClient, client := newMockedClient(mock.GetDefaultClassicationResponseHappy, http.StatusOK)

		Convey("When GetDefaultClassification is called", func() {
			req := cantabularmetadata.GetDefaultClassificationRequest{
				Dataset:   "test_dataset",
				Variables: []string{"test_variable_1", "test_variable_2"},
			}

			resp, err := client.GetDefaultClassification(ctx, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular metadata service", func() {
				So(httpClient.PostCalls(), ShouldHaveLength, 1)
				So(httpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.metadata.host/graphql")
				validateQuery(
					httpClient.PostCalls()[0].Body,
					cantabularmetadata.QueryDefaultClassification,
					cantabularmetadata.QueryData{
						Dataset:   "test_dataset",
						Variables: []string{"test_variable_1", "test_variable_2"},
					},
				)
			})

			expected := &cantabularmetadata.GetDefaultClassificationResponse{
				Variables: []string{"test_variable_2"},
			}

			Convey("And the expected response is returned", func() {
				So(resp, ShouldResemble, expected)
			})
		})
	})
}

func TestGetDefaultClassificationNoDefaultVariables(t *testing.T) {
	Convey("Given a correct GetDefaultClassication response from the /graphql endpoint", t, func() {
		ctx := context.Background()
		httpClient, client := newMockedClient(mock.GetDefaultClassicationResponseNoDefaultVariables, http.StatusOK)

		Convey("When GetDefaultClassification is called", func() {
			req := cantabularmetadata.GetDefaultClassificationRequest{
				Dataset:   "test_dataset",
				Variables: []string{"test_variable_1", "test_variable_2"},
			}

			resp, err := client.GetDefaultClassification(ctx, req)

			Convey("Then the experected error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular metadata service", func() {
				So(httpClient.PostCalls(), ShouldHaveLength, 1)
				So(httpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.metadata.host/graphql")
				validateQuery(
					httpClient.PostCalls()[0].Body,
					cantabularmetadata.QueryDefaultClassification,
					cantabularmetadata.QueryData{
						Dataset:   "test_dataset",
						Variables: []string{"test_variable_1", "test_variable_2"},
					},
				)
			})

			expected := &cantabularmetadata.GetDefaultClassificationResponse{}

			Convey("And the expected response should be returned", func() {
				So(resp, ShouldResemble, expected)
			})
		})
	})
}

func TestGetDefaultClassificationMultipleDefaultVariables(t *testing.T) {
	Convey("Given a correct GetDefaultClassication response from the /graphql endpoint", t, func() {
		ctx := context.Background()
		httpClient, client := newMockedClient(mock.GetDefaultClassicationResponseMultipleDefaultVariables, http.StatusOK)

		Convey("When GetDefaultClassification is called", func() {
			req := cantabularmetadata.GetDefaultClassificationRequest{
				Dataset:   "test_dataset",
				Variables: []string{"test_variable_1", "test_variable_2"},
			}

			resp, err := client.GetDefaultClassification(ctx, req)

			Convey("Then the experected error should be nil", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular metadata service", func() {
				So(httpClient.PostCalls(), ShouldHaveLength, 1)
				So(httpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.metadata.host/graphql")
				validateQuery(
					httpClient.PostCalls()[0].Body,
					cantabularmetadata.QueryDefaultClassification,
					cantabularmetadata.QueryData{
						Dataset:   "test_dataset",
						Variables: []string{"test_variable_1", "test_variable_2"},
					},
				)
			})

			expected := &cantabularmetadata.GetDefaultClassificationResponse{
				Variables: []string{"test_variable_1", "test_variable_2"},
			}
			Convey("And the response should be nil", func() {
				So(resp, ShouldResemble, expected)
			})
		})
	})
}

func TestGetDefaultClassificationResponseCantabularError(t *testing.T) {
	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		ctx := context.Background()
		_, client := newMockedClient(mock.ErrorResponseNoDataset, http.StatusOK)

		Convey("When GetAllDimensions is called", func() {
			req := cantabularmetadata.GetDefaultClassificationRequest{
				Dataset:   "test_dataset",
				Variables: []string{"test_variable_1", "test_variable_2"},
			}
			resp, err := client.GetDefaultClassification(ctx, req)

			Convey("Then the expected error is returned", func() {
				So(client.StatusCode(err), ShouldNotBeNil)
				// TODO: Talk to SCC ask for consistent reporting of error codes
				// in errors
				So(client.StatusCode(err), ShouldResemble, http.StatusBadGateway)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}
