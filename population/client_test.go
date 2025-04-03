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

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
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

func TestGetAreaTypes(t *testing.T) {
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreaTypesInput{
			PopulationType: "test",
		}

		client.GetAreaTypes(context.Background(), input)

		Convey("it should call the area types endpoint, serializing the dataset query", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/test/area-types?limit=0&offset=0")
		})
	})

	Convey("Given authentication tokens", t, func() {
		const userAuthToken = "user"
		const serviceAuthToken = "service"

		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreaTypesInput{
			AuthTokens: AuthTokens{
				ServiceAuthToken: serviceAuthToken,
				UserAuthToken:    userAuthToken,
			},
			PopulationType: "test",
		}

		client.GetAreaTypes(context.Background(), input)

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

		input := GetAreaTypesInput{
			PopulationType: "test",
		}
		types, err := client.GetAreaTypes(context.Background(), input)

		Convey("it should return a list of area types", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, areaTypes)
		})
	})

	Convey("Given the population types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetAreaTypesInput{
			PopulationType: "test",
		}
		_, err := client.GetAreaTypes(context.Background(), input)

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

		input := GetAreaTypesInput{
			PopulationType: "test",
		}
		_, err := client.GetAreaTypes(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "area-types" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaTypesInput{
			PopulationType: "test",
		}
		_, err := client.GetAreaTypes(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreaTypesInput{
			PopulationType: "test",
		}
		_, err := client.GetAreaTypes(nil, input)

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
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "testText",
		}
		client.GetAreas(context.Background(), input)

		Convey("it should call the areas endpoint, serializing the dataset, area type and text query params", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/testDataSet/area-types/testAreaType/areas?limit=0&offset=0&q=testText")
		})
	})

	Convey("Given a valid request with an empty text param", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreasInput{
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
		}

		client.GetAreas(context.Background(), input)

		Convey("it should call the areas endpoint, omitting the text query param", func() {
			calls := stubClient.DoCalls()

			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/testDataSet/area-types/testAreaType/areas?limit=0&offset=0")
		})
	})

	Convey("Given authentication tokens", t, func() {
		const userAuthToken = "user"
		const serviceAuthToken = "service"

		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreasInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
		}

		client.GetAreas(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid areas response payload", t, func() {
		areas := GetAreasResponse{
			PaginationResponse: PaginationResponse{
				PaginationParams: PaginationParams{
					Limit:  2,
					Offset: 0,
				},
				Count:      2,
				TotalCount: 6,
			},
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
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
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
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
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
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreasInput{
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
		}
		_, err := client.GetAreas(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreasInput{
			AuthTokens:     AuthTokens{},
			PopulationType: "testDataSet",
			AreaTypeID:     "testAreaType",
			Text:           "",
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
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: "popType",
			AreaType:       "areaType",
			Area:           "ID",
		}

		client.GetArea(context.Background(), input)
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
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: "popType",
			AreaType:       "areaType",
			Area:           "ID",
		}

		_, err := client.GetArea(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})

	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreaInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
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
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
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

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		client.GetPopulationTypes(context.Background(), input)

		Convey("it should call the population types endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types?limit=0&offset=0")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}

		client.GetPopulationTypes(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid population types response payload", t, func() {
		ptypes := GetPopulationTypesResponse{
			Items: []PopulationType{{Name: "test", Label: "test", Description: "Test", Type: "test"}},
		}

		resp, err := json.Marshal(ptypes)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		types, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, ptypes)
		})
	})

	Convey("Given the population types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
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

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationTypes(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetPopulationTypesInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationTypes(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetPopulationType(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"

	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: "test",
		}
		client.GetPopulationType(context.Background(), input)

		Convey("it should call the population types endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/test")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}

		client.GetPopulationType(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid population types response payload", t, func() {
		ptypes := GetPopulationTypeResponse{
			PopulationType: PopulationType{Name: "test", Label: "test", Description: "Test", Type: "test"},
		}

		resp, err := json.Marshal(ptypes)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: "test",
		}
		types, err := client.GetPopulationType(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(types, ShouldResemble, ptypes)
		})
	})

	Convey("Given the population types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationType(context.Background(), input)

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

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationType(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationType(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetPopulationTypeInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetPopulationType(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetAreaTypesParent(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "populationType"
	const areaTypeId = "areaId"
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetAreaTypeParentsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			AreaTypeID:     areaTypeId,
		}
		client.GetAreaTypeParents(context.Background(), input)

		Convey("it should call the area types parens endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/v1/population-types/populationType/area-types/areaId/parents?limit=0&offset=0")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}

		client.GetAreaTypeParents(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid areaTypes parents response payload", t, func() {
		parents := GetAreaTypeParentsResponse{
			PaginationResponse: PaginationResponse{
				PaginationParams: PaginationParams{
					Limit:  2,
					Offset: 0,
				},
				Count:      2,
				TotalCount: 6,
			},
			AreaTypes: []AreaType{{ID: "test", Label: "Test", TotalCount: 2}},
		}

		resp, err := json.Marshal(parents)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
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
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
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
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetAreaTypeParentsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetAreaTypeParents(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetAreaTypeParentsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetAreaTypeParents(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetParentAreaCount(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "populationType"
	const areaTypeId = "areaId"
	const parentAreaTypeId = "parentAreaTypeId"
	areas := []string{"area1", "area2"}
	svar := "var1"

	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		client.GetParentAreaCount(context.Background(), input)

		Convey("it should call the parent areas count endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(
				calls[0].Req.URL.String(),
				ShouldEqual,
				"http://test.test:2000/v1/population-types/populationType/area-types/areaId/parents/parentAreaTypeId/areas-count?areas=area1%2Carea2&svar=var1")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}

		client.GetParentAreaCount(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid parents areas count response payload", t, func() {
		resp, err := json.Marshal(1)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		res, err := client.GetParentAreaCount(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, 1)
		})
	})

	Convey("Given the population API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		_, err := client.GetParentAreaCount(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the parents area count endpoint returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		_, err := client.GetParentAreaCount(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		_, err := client.GetParentAreaCount(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be processed", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		_, err := client.GetParentAreaCount(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})

	Convey("Given the parent areas count request response cannot be converted to int", t, func() {
		resp, err := json.Marshal("some incorrect value")
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetParentAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType:   populationType,
			AreaTypeID:       areaTypeId,
			ParentAreaTypeID: parentAreaTypeId,
			Areas:            areas,
			SVarID:           svar,
		}
		_, err = client.GetParentAreaCount(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})
}

func TestGetBlockedAreaCount(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "populationType"
	areas := []string{"area1", "area2"}
	const svar = "var1"

	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should call the blocked areas count endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(
				calls[0].Req.URL.String(),
				ShouldEqual,
				"http://test.test:2000/v1/population-types/populationType/blocked-areas-count?areas=area1%2Carea2&fvar=var1&vars=var1")
		})
	})

	Convey("Given authentication tokens", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}

		client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should set the auth headers on the request", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
		})
	})

	Convey("Given a valid blocked areas count response payload", t, func() {

		result := cantabular.GetBlockedAreaCountResult{
			Passed:     1,
			Blocked:    2,
			Total:      3,
			TableError: "",
		}
		resp, err := json.Marshal(result)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}

		res, err := client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, &result)
		})
	})

	Convey("Given a areas count result where quer doesn't fail but table level error is returned", t, func() {
		result := cantabular.GetBlockedAreaCountResult{
			Passed:     0,
			Blocked:    0,
			Total:      0,
			TableError: "some error at table level",
		}
		resp, err := json.Marshal(result)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}

		res, err := client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, &result)
		})
	})

	Convey("Given the population API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		_, err := client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the blocked area count endpoint returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		_, err := client.GetBlockedAreaCount(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "areas" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		_, err := client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be processed", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		_, err := client.GetBlockedAreaCount(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})

	Convey("Given the parent areas count request response cannot be converted to int", t, func() {
		resp, err := json.Marshal("some incorrect value")
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetBlockedAreaCountInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			Variables:      []string{svar},
			Filter:         Filter{Variable: svar, Codes: areas},
		}
		_, err = client.GetBlockedAreaCount(context.Background(), input)

		Convey("it should return a list of population types", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
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
// Usage:
// `So(err, shouldBeDPError)`
// `So(err, shouldBeDPError, 404)`
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
