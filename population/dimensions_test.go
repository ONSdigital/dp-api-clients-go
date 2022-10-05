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

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetDimensions(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "populationId"
	const SearchString = "searchString"
	Convey("Given a valid request", t, func() {
		stubClient := newStubClient(&http.Response{Body: ioutil.NopCloser(bytes.NewReader(nil))}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
			PopulationType: populationType,
			SearchString:   SearchString,
		}

		client.GetDimensions(context.Background(), input)

		Convey("it should call the get dimensions endpoint", func() {
			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			fmt.Println(calls[0].Req.URL.String())
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/population-types/populationId/dimensions?limit=0&offset=0&q=searchString")
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

	Convey("Given a valid get dimensions payload", t, func() {
		dimensions := GetDimensionsResponse{
			PaginationResponse: PaginationResponse{
				PaginationParams: PaginationParams{
					Limit:  2,
					Offset: 0,
				},
				Count:      2,
				TotalCount: 6,
			},
			Dimensions: []Dimension{
				{
					Name:       "",
					Label:      "Accommodation type (8 categories)",
					TotalCount: 8,
				},
				{
					Name:       "",
					Label:      "Type of central heating in household (13 categories)",
					TotalCount: 13,
				}},
		}

		resp, err := json.Marshal(dimensions)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client := newHealthClient(stubClient)

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		res, err := client.GetDimensions(context.Background(), input)

		Convey("it should return a list of dimensions", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, dimensions)
		})
	})

	Convey("Given the get dimensions API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetDimensions(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the get dimensions API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetDimensions(context.Background(), input)

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
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "dimensions" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetDimensions(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetDimensionsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetDimensions(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetCategorisations(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "population-id"
	const dimensionID = "dimension-id"

	Convey("Given a valid request categorisations payload", t, func() {
		categorisations := GetCategorisationsResponse{
			PaginationResponse: PaginationResponse{
				PaginationParams: PaginationParams{
					Limit:  2,
					Offset: 0,
				},
				Count:      2,
				TotalCount: 6,
			},
			Items: []Dimension{
				{
					Name:       "",
					Label:      "Accommodation type (8 categories)",
					TotalCount: 8,
				},
				{
					Name:       "",
					Label:      "Accomodation type (13 categories)",
					TotalCount: 13,
				}},
		}

		resp, err := json.Marshal(categorisations)
		So(err, ShouldBeNil)

		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp)),
		}, nil)
		client, err := NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000/v1", stubClient))
		So(err, ShouldBeNil)

		Convey("Given a valid request with authentication tokens", func() {
			input := GetCategorisationsInput{
				AuthTokens: AuthTokens{
					UserAuthToken:    userAuthToken,
					ServiceAuthToken: serviceAuthToken,
				},
				PaginationParams: PaginationParams{
					Limit:  10,
					Offset: 0,
				},
				PopulationType: populationType,
				Dimension:      dimensionID,
			}

			res, err := client.GetCategorisations(context.Background(), input)
			So(err, ShouldBeNil)

			Convey("Then the auth headers should be set on the request", func() {
				calls := stubClient.DoCalls()
				So(calls, ShouldNotBeEmpty)
				So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
			})

			Convey("Then the GetCategorisations endpoint should be called", func() {
				calls := stubClient.DoCalls()
				So(calls, ShouldNotBeEmpty)
				fmt.Println(calls[0].Req.URL.String())
				So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/population-types/population-id/dimensions/dimension-id/categorisations?limit=10&offset=0")
			})

			Convey("And A list of categorisations should be returned", func() {
				So(err, ShouldBeNil)
				So(res, ShouldResemble, categorisations)
			})
		})
	})

	Convey("Given the get population-types API returns an error", t, func() {
		stubClient := newStubClient(nil, errors.New("oh no"))

		client := newHealthClient(stubClient)

		input := GetCategorisationsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetCategorisations(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the get dimensions API returns a status code of 404", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "errors": ["not found"] }`))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetCategorisationsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetCategorisations(context.Background(), input)

		Convey("the error chain should contain the original Errors type", func() {
			So(err, shouldBeDPError, http.StatusNotFound)

			var respErr ErrorResp
			ok := errors.As(err, &respErr)
			t.Log("DEBUG", err)
			So(ok, ShouldBeTrue)
			So(respErr, ShouldResemble, ErrorResp{Errors: []string{"not found"}})
		})
	})

	Convey("Given the response cannot be deserialized", t, func() {
		stubClient := newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "dimensions" `))),
		}, nil)

		client := newHealthClient(stubClient)

		input := GetCategorisationsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetCategorisations(context.Background(), input)

		Convey("it should return an internal error", func() {
			So(err, shouldBeDPError, http.StatusInternalServerError)
		})
	})

	Convey("Given the request cannot be created", t, func() {
		client := newHealthClient(newStubClient(nil, nil))

		input := GetCategorisationsInput{
			AuthTokens: AuthTokens{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		_, err := client.GetCategorisations(nil, input)

		Convey("it should return a client error", func() {
			So(err, shouldBeDPError, http.StatusBadRequest)
		})
	})
}

func TestGetBaseVariable(t *testing.T) {
	const userAuthToken = "user"
	const serviceAuthToken = "service"
	const populationType = "population-id"
	const variable = "variable"

	var stubClient *dphttp.ClienterMock
	var client *Client

	expectedURL := fmt.Sprintf("http://test.test:2000/population-types/%s/dimensions/%s/base", populationType, variable)
	input := GetBaseVariableInput{
		AuthTokens: AuthTokens{
			UserAuthToken:    userAuthToken,
			ServiceAuthToken: serviceAuthToken,
		},

		PopulationType: populationType,
		Variable:       variable,
	}
	Convey("Creating a valid client", t, func() {
		stubClient = newStubClient(&http.Response{
			Body: ioutil.NopCloser(bytes.NewReader(nil))},
			nil,
		)

		var err error
		client, err = NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000", stubClient))
		So(err, ShouldBeNil)
		Convey("Given a valid request it should call the base variable endpoint, serializing the dataset query", func() {
			client.GetBaseVariable(context.Background(), input)

			calls := stubClient.DoCalls()
			So(calls, ShouldNotBeEmpty)

			So(calls[0].Req.URL.String(), ShouldEqual, expectedURL)

		})

		Convey("Given a valid request with authentication tokens", func() {
			client.GetBaseVariable(context.Background(), input)
			Convey("Then the auth headers should be set on the request", func() {
				calls := stubClient.DoCalls()
				So(calls, ShouldNotBeEmpty)
				So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken)
			})
		})
	})

	Convey("Given a valid request categorisations payload", t, func() {
		baseVariable := GetBaseVariableResponse{
			Name:  "givenName",
			Label: "givenLabel",
		}

		resp, err := json.Marshal(baseVariable)
		So(err, ShouldBeNil)
		stubClient = newStubClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(resp))},
			nil,
		)
		client, err = NewWithHealthClient(health.NewClientWithClienter("", "http://test.test:2000", stubClient))
		So(err, ShouldBeNil)

		res, err := client.GetBaseVariable(context.Background(), input)

		Convey("And base variable should be returned", func() {
			So(err, ShouldBeNil)
			So(res, ShouldResemble, baseVariable)
		})
	})

}
