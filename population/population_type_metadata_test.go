package population

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDefaultDatasetMetadata(t *testing.T) {
	const userAuthToken = "user"
	const populationType = "UR"

	Convey("Given a valid metadata request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetPopulationTypeMetadataInput{
			AuthTokens: AuthTokens{
				UserAuthToken: userAuthToken,
			},
			PopulationType: populationType,
		}
		client.GetPopulationTypeMetadata(context.Background(), input)
		Convey("It should call the population types endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/population-types/UR/metadata")
		})

	})

	Convey("Given a valid payload", t, func() {
		response := GetPopulationTypeMetadataResponse{
			PopulationType:   "UR",
			DefaultDatasetID: "defaultID",
			Edition:          "2021",
			Version:          1,
		}

		resp, err := json.Marshal(response)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetPopulationTypeMetadataInput{
			AuthTokens: AuthTokens{
				UserAuthToken: userAuthToken,
			},
			PopulationType: populationType,
		}
		res, err := client.GetPopulationTypeMetadata(context.Background(), input)

		Convey("it should return a list of dimensions", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, response)
		})
	})

	Convey("Given the get metadata API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetPopulationTypeMetadataInput{
			AuthTokens: AuthTokens{
				UserAuthToken: userAuthToken,
			},
			PopulationType: populationType,
		}
		_, err := client.GetPopulationTypeMetadata(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the get metadata API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetPopulationTypeMetadataInput{
			AuthTokens: AuthTokens{
				UserAuthToken: userAuthToken,
			},
			PopulationType: populationType,
		}
		_, err := client.GetPopulationTypeMetadata(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})
}
