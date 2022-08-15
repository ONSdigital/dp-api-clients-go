package population

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/http"
)

func TestNewClient(t *testing.T) {
	const invalidURL = "a#$%^&*(url$#$%%^("

	Convey("Given NewClient is passed an invalid URL", t, func() {
		_, err := NewClient(invalidURL)

		Convey("the constructor should return an error", func() {
			So(err, ShouldBeError)
		})
	})

	Convey("Given NewWithHealthClient is passed an invalid URL", t, func() {
		_, err := NewWithHealthClient(health.NewClientWithClienter("", invalidURL, newStubClient(nil, nil)))

		Convey("the constructor should return an error", func() {
			So(err, ShouldBeError)
		})
	})
}

func TestGetPopulationAreaTypes(t *testing.T) {
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		_, _ = client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("it should call the area types endpoint, serializing the dataset query", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/test/area-types")
		})
	})

	Convey("Given authentication tokens", t, func() {
		const userAuthToken = "user"
		const serviceAuthToken = "service"

		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		_, _ = client.GetPopulationAreaTypes(context.Background(), userAuthToken, serviceAuthToken, "test")

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid area types response payload", t, func() {
		areaTypes := GetAreaTypesResponse{
			AreaTypes: []AreaType{{ID: "test", Label: "Test", TotalCount: 5}},
		}

		resp, err := json.Marshal(areaTypes)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)

		client := newHealthClient(stubClient)

		types, err := client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("it should return a list of area types", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, areaTypes)
		})
	})

	Convey("Given the population types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		_, err := client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the population types API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		_, err := client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the population types API returns a status code other than 200/400", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "area-types": [] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		_, err := client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "area-types" `))),
		}, nil)

		client := newHealthClient(stubClient)

		_, err := client.GetPopulationAreaTypes(context.Background(), "", "", "test")

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		_, err := client.GetPopulationAreaTypes(nil, "", "", "test")

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetAreas(t *testing.T) {
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "testText",
		}
		_, _ = client.GetAreas(context.Background(), input)

		Convey("it should call the areas endpoint, serializing the dataset, area type and text query params", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/testDataSet/area-types/testAreaType/areas?q=testText")
		})
	})

	Convey("Given a valid request with an empty text param", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}

		_, _ = client.GetAreas(context.Background(), input)

		Convey("it should call the areas endpoint, omitting the text query param", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/testDataSet/area-types/testAreaType/areas")
		})
	})

	Convey("Given authentication tokens", t, func() {
		const userAuthToken = "user"
		const serviceAuthToken = "service"

		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}

		_, _ = client.GetAreas(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid areas response payload", t, func() {
		areas := GetAreasResponse{
			Areas: []Area{{ID: "test", Label: "Test", AreaType: "city"}},
		}

		resp, err := json.Marshal(areas)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		types, err := client.GetAreas(context.Background(), input)

		Convey("it should return a list of areas", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, areas)
		})
	})

	Convey("Given the dimensions API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the dimensions API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the dimensions API returns a status code other than 200/400", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas": [] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreasInput{
			UserAuthToken:    "",
			ServiceAuthToken: "",
			DatasetID:        "testDataSet",
			AreaTypeID:       "testAreaType",
			Text:             "",
		}
		_, err := client.GetAreas(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetArea(t *testing.T) {

	const userAuthToken = "user"
	const serviceAuthToken = "service"

	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreaInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
			PopulationType:   "popType",
			AreaType:         "areaType",
			Area:             "ID",
		}

		_, _ = client.GetArea(context.Background(), input)
		calls := stubClient.DoCalls()

		Convey("it should call the specific area endpoint", func() {
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/popType/area-types/areaType/areas/ID")
		})

		Convey("it should set the auth headers on the request", func() {
			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})

	})

	Convey("Given the population types api returns a 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)
		input := GetAreaInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
			PopulationType:   "popType",
			AreaType:         "areaType",
			Area:             "ID",
		}

		_, err := client.GetArea(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})

	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreaInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetArea(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areasdjfhas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetArea(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

}

func TestGetPopulationTypes(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"

	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, _ = client.GetPopulationTypes(context.Background(), input)

		Convey("it should call the population types endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}

		_, _ = client.GetPopulationTypes(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid areas response payload", t, func() {
		areas := GetAreasResponse{
			Areas: []Area{{ID: "test", Label: "Test", AreaType: "city"}},
		}

		resp, err := json.Marshal(areas)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		types, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, areas)
		})
	})

	Convey("Given the population types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the population types API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the population types API returns a status code other than 200/400", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas": [] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreasInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetPopulationTypes(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetAreaTypesParent(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const datasetId = "datasetId"
	const areaTypeId = "areaId"
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
			DatasetID:        datasetId,
			AreaTypeID:       areaTypeId,
		}
		_, _ = client.GetAreaTypeParents(context.Background(), input)

		Convey("it should call the area types parens endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/datasetId/area-types/areaId/parents")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}

		_, _ = client.GetAreaTypeParents(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid areaTypes parents response payload", t, func() {
		parents := GetAreaTypeParentsResponse{
			AreaTypes: []AreaTypes{{ID: "test", Label: "Test", TotalCount: 2}},
		}

		resp, err := json.Marshal(parents)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		res, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, parents)
		})
	})

	Convey("Given the area types parents API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the area types parents API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the area types parents API returns a status code other than 200/400", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas": [] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreaTypeParentsInput{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		}
		_, err := client.GetAreaTypeParents(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

// newHealthClient creates a new Client from an existing Clienter
func newHealthClient(client dphttp.Clienter) *Client {
	stubClientWithHealth := health.NewClientWithClienter("", "", client)
	healthClient, err := NewWithHealthClient(stubClientWithHealth)
	if err != nil {
		panic(err)
	}

	return healthClient
}

// newStubClient creates a stub Clienter which always responds to `Do` with the
// provided response/error.
func newStubClient(response *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(_ context.Context, _ *http.Request) (*http.Response, error) {
			return response, err
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

// shouldBeDPError is a GoConvey matcher that asserts the passed in error
// includes a dperrors.Error within the chain, and optionally that the status
// code matches the expected value.
// Usage:`So(err, shouldBeDPError)`
//       `So(err, shouldBeDPError, 404)`
func shouldBeDPError(actual interface{}, expected ...interface{}) string {
	err, ok := actual.(error)
	if !ok {
		return "expected to find error"
	}

	var dpErr *dperrors.Error
	if ok := errors.As(err, &dpErr); !ok {
		return "did not find dperrors.Error in the chain"
	}

	if len(expected) == 0 {
		return ""
	}

	statusCode, ok := expected[0].(int)
	if !ok {
		return "status code could not be parsed"
	}

	if statusCode != dpErr.Code() {
		return fmt.Sprintf("expected status code %d, got %d", statusCode, dpErr.Code())
	}

	return ""
}

// shouldHaveAuthHeaders is a GoConvey matcher that asserts the values of the
// auth headers on a request match the expected values.
// Usage: `So(request, shouldHaveAuthHeaders, "userToken", "serviceToken")`
func shouldHaveAuthHeaders(actual interface{}, expected ...interface{}) string {
	req, ok := actual.(*http.Request)
	if !ok {
		return "expected to find http.Request"
	}

	if len(expected) != 2 {
		return "expected a user header and a service header"
	}

	expUserHeader, ok := expected[0].(string)
	if !ok {
		return "user header must be a string"
	}

	expSvcHeader, ok := expected[1].(string)
	if !ok {
		return "service header must be a string"
	}

	florenceToken := req.Header.Get("X-Florence-Token")
	if florenceToken != expUserHeader {
		return fmt.Sprintf("expected X-Florence-Token value %s, got %s", florenceToken, expUserHeader)
	}

	svcHeader := req.Header.Get("Authorization")
	if svcHeader != fmt.Sprintf("Bearer %s", expSvcHeader) {
		return fmt.Sprintf("expected Authorization value %s, got %s", svcHeader, expSvcHeader)
	}

	return ""
}
