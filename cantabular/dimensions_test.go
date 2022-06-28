package cantabular_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	dphttp "github.com/ONSdigital/dp-net/http"
)

func TestGetDimensionsHappy(t *testing.T) {
	Convey("Given a correct getDimensions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyGetDimensions, http.StatusOK)

		Convey("When GetDimensions is called", func() {
			resp, err := cantabularClient.GetDimensions(testCtx, "Teaching-Dataset")

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryDimensions,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensions)
			})
		})
	})
}

func TestGetDimensionsUnhappy(t *testing.T) {
	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When GetDimensions is called", func() {
			resp, err := cantabularClient.GetDimensions(testCtx, "InexistentDataset")

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetDimensions is called", func() {
			resp, err := cantabularClient.GetDimensions(testCtx, "Teaching-Dataset")

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedInternalServeError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestGetGeographyDimensionsHappy(t *testing.T) {
	Convey("Given a correct getGeographyDimensions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyGetGeographyDimensions, http.StatusOK)

		Convey("When GetGeographyDimensions is called", func() {
			resp, err := cantabularClient.GetGeographyDimensions(testCtx, cantabular.GetGeographyDimensionsRequest{
				Dataset: "Teaching-Dataset",
				PaginationParams: cantabular.PaginationParams{
					Limit:  10,
					Offset: 0,
				},
			})

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryGeographyDimensions,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
						PaginationParams: cantabular.PaginationParams{
							Limit:  10,
							Offset: 0,
						},
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedGeographyDimensions)
			})
		})
	})
}

func TestGetGeographyDimensionsUnhappy(t *testing.T) {
	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When GetGeographyDimensions is called", func() {
			resp, err := cantabularClient.GetGeographyDimensions(testCtx, cantabular.GetGeographyDimensionsRequest{
				Dataset: "InexistentDataset",
				PaginationParams: cantabular.PaginationParams{
					Limit:  10,
					Offset: 0,
				},
			})

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetGeographyDimensions is called", func() {
			resp, err := cantabularClient.GetGeographyDimensions(testCtx, cantabular.GetGeographyDimensionsRequest{
				Dataset: "Teaching-Dataset",
				PaginationParams: cantabular.PaginationParams{
					Limit:  10,
					Offset: 0,
				},
			})

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedInternalServeError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestGetDimensionsByNameHappy(t *testing.T) {
	Convey("Given a correct getDimensionsByName response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyGetDimensionsByName, http.StatusOK)

		Convey("When GetDimensionsByName is called", func() {
			resp, err := cantabularClient.GetDimensionsByName(testCtx, cantabular.GetDimensionsByNameRequest{
				Dataset: "Teaching-Dataset",
				DimensionNames: []cantabular.DimensionsOptions{
					{ID: "Age"},
					{ID: "Region"},
				},
			})

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryDimensionsByName,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
						Variables: []cantabular.DimensionsOptions{
							{ID: "Age"},
							{ID: "Region"},
						},
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensionsByName)
			})
		})
	})
}

func TestGetDimensionsByNameUnhappy(t *testing.T) {
	testCtx := context.Background()

	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When the GetDimensionsByName method is called", func() {
			req := cantabular.GetDimensionsByNameRequest{
				Dataset: "InexistentDataset",
				DimensionNames: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "Occupation"},
				},
			}
			resp, err := cantabularClient.GetDimensionsByName(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)

			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a no-variable graphql error response from the /graphql endpoint", t, func() {
		_, cantabularClient := newMockedClient(mockRespBodyNoVariable, http.StatusOK)

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.GetDimensionsByNameRequest{
				Dataset: "Teaching-Dataset",
				DimensionNames: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "inexistentVariable"},
				},
			}
			resp, err := cantabularClient.GetDimensionsByName(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusBadRequest)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetDimensionsByName is called", func() {
			req := cantabular.GetDimensionsByNameRequest{
				Dataset: "Teaching-Dataset",
				DimensionNames: []cantabular.DimensionsOptions{
					{ID: "Age"},
					{ID: "Region"},
				},
			}
			resp, err := cantabularClient.GetDimensionsByName(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedInternalServeError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestSearchDimensionsHappy(t *testing.T) {
	Convey("Given a correct searchDimensions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodySearchDimensions, http.StatusOK)

		Convey("When SearchDimensions is called", func() {
			resp, err := cantabularClient.SearchDimensions(testCtx, cantabular.SearchDimensionsRequest{
				Dataset: "Teaching-Dataset",
				Text:    "country",
			})

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryDimensionsSearch,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
						Text:    "country",
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedSearchDimensionsResponse)
			})
		})
	})
}

func TestSearchDimensionsUnhappy(t *testing.T) {
	testCtx := context.Background()

	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When the SearchDimensions method is called", func() {
			req := cantabular.SearchDimensionsRequest{
				Dataset: "InexistentDataset",
				Text:    "country",
			}
			resp, err := cantabularClient.SearchDimensions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When SearchDimensions is called", func() {
			req := cantabular.SearchDimensionsRequest{
				Dataset: "Teaching-Dataset",
				Text:    "country",
			}
			resp, err := cantabularClient.SearchDimensions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedInternalServeError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestGetDimensionOptionsHappy(t *testing.T) {
	Convey("Given a correct getDimensionOptions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyGetDimensionOptions, http.StatusOK)

		Convey("When GetDimensionOptions is called", func() {
			req := cantabular.GetDimensionOptionsRequest{
				Dataset: "Teaching-Dataset",
				Dimensions: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "Occupation"},
				},
				Filters: []cantabular.Filter{{Variable: "Country", Codes: []string{"E", "W"}}},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryDimensionOptions,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
						Variables: []cantabular.DimensionsOptions{
							{ID: "Country"},
							{ID: "Age"},
							{ID: "Occupation"},
						},
						Filters: []cantabular.Filter{{Variable: "Country", Codes: []string{"E", "W"}}},
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensionOptions)
			})
		})

		Convey("When GetDimensionOptions is called without filters", func() {
			req := cantabular.GetDimensionOptionsRequest{
				Dataset: "Teaching-Dataset",
				Dimensions: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "Occupation"},
				},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryDimensionOptions,
					cantabular.QueryData{
						Dataset: "Teaching-Dataset",
						Variables: []cantabular.DimensionsOptions{
							{ID: "Country"},
							{ID: "Age"},
							{ID: "Occupation"},
						},
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensionOptions)
			})
		})
	})
}

func TestGetDimensionOptionsUnhappy(t *testing.T) {
	testCtx := context.Background()

	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.GetDimensionOptionsRequest{
				Dataset: "InexistentDataset",
				Dimensions: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "Occupation"},
				},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a no-variable graphql error response from the /graphql endpoint", t, func() {
		_, cantabularClient := newMockedClient(mockRespBodyNoVariable, http.StatusOK)

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.GetDimensionOptionsRequest{
				Dataset: "Teaching-Dataset",
				Dimensions: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "inexistentVariable"},
				},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusBadRequest)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetDimensionOptions is called", func() {
			req := cantabular.GetDimensionOptionsRequest{
				Dataset: "Teaching-Dataset",
				Dimensions: []cantabular.DimensionsOptions{
					{ID: "Country"},
					{ID: "Age"},
					{ID: "Occupation"},
				},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, expectedInternalServeError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

func TestGetAreas(t *testing.T) {
	Convey("Given a valid response from the /graphql endpoint", t, func() {
		const dataset = "Example"
		const variable = "city"
		const category = "london"

		testCtx := context.Background()
		mockHttpClient, cantabularClient := newMockedClient(mockRespBodyGetAreas, http.StatusOK)

		Convey("When GetAreas is called", func() {
			req := cantabular.GetAreasRequest{
				Dataset:  dataset,
				Variable: variable,
				Category: category,
			}

			resp, err := cantabularClient.GetAreas(testCtx, req)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected query is posted to cantabular api-ext", func() {
				So(mockHttpClient.PostCalls(), ShouldHaveLength, 1)
				So(mockHttpClient.PostCalls()[0].URL, ShouldEqual, "cantabular.ext.host/graphql")
				validateQuery(
					mockHttpClient.PostCalls()[0].Body,
					cantabular.QueryAreas,
					cantabular.QueryData{
						Dataset:  dataset,
						Text:     variable,
						Category: category,
					},
				)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedAreas)
			})
		})
	})
}

func TestGetAreasUnhappy(t *testing.T) {
	stubReq := cantabular.GetAreasRequest{
		Dataset:  "Example",
		Variable: "city",
		Category: "london",
	}

	Convey("Given a no-dataset graphql error response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespBodyNoDataset, http.StatusOK)

		Convey("When GetAreas is called", func() {

			resp, err := cantabularClient.GetAreas(testCtx, stubReq)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusNotFound)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a 500 HTTP Status response from the /graphql endpoint", t, func() {
		testCtx := context.Background()
		_, cantabularClient := newMockedClient(mockRespInternalServerErr, http.StatusInternalServerError)

		Convey("When GetAreas is called", func() {
			resp, err := cantabularClient.GetAreas(testCtx, stubReq)

			Convey("Then the expected error is returned", func() {
				So(cantabularClient.StatusCode(err), ShouldResemble, http.StatusInternalServerError)
			})

			Convey("And no response is returned", func() {
				So(resp, ShouldBeNil)
			})
		})
	})
}

// newMockedClient creates a new cantabular client with a mocked response for post requests,
// according to the provided response string and status code.
func newMockedClient(response string, statusCode int) (*dphttp.ClienterMock, *cantabular.Client) {
	mockHttpClient := &dphttp.ClienterMock{
		PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
			return Response(
				[]byte(response),
				statusCode,
			), nil
		},
	}

	cantabularClient := cantabular.NewClient(
		cantabular.Config{
			Host:       "cantabular.host",
			ExtApiHost: "cantabular.ext.host",
		},
		mockHttpClient,
		nil,
	)

	return mockHttpClient, cantabularClient
}

// mockRespBodyGetDimensions is a successful 'get dimensions' query respose that is returned from a mocked client for testing
var mockRespBodyGetDimensions = `
{
	"data": {
		"dataset": {
			"variables": {
				"edges": [
					{
						"node": {
							"categories": {
								"totalCount":8
							},
							"label": "Age",
							"mapFrom": [],
							"name": "Age"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount":2
							},
							"label": "Country",
							"mapFrom": [
								{
									"edges": [
										{
											"node": {
												"filterOnly": "false",
												"label": "Region",
												"name": "Region"
											}
										}
									]
								}
							],
							"name": "Country"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount": 6
							},
							"label": "Health",
							"mapFrom": [],
							"name": "Health"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount":5
							},
							"label": "Marital Status",
							"mapFrom": [],
							"name": "Marital Status"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount":10
							},
							"label": "Region",
							"mapFrom": [],
							"name": "Region"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount":2
							},
							"label": "Sex",
							"mapFrom":[],
							"name":"Sex"
						}
					}
				]
			}
		}
	}
}`

// expectedDimensions is the expected response struct generated from a successful 'get dimensions' query for testing
var expectedDimensions = cantabular.GetDimensionsResponse{
	Dataset: gql.DatasetVariables{
		Variables: gql.Variables{
			Edges: []gql.Edge{
				{
					Node: gql.Node{
						Name:       "Age",
						Label:      "Age",
						Categories: gql.Categories{TotalCount: 8},
						MapFrom:    []gql.Variables{},
					},
				},
				{
					Node: gql.Node{
						Name:       "Country",
						Label:      "Country",
						Categories: gql.Categories{TotalCount: 2},
						MapFrom: []gql.Variables{
							{
								Edges: []gql.Edge{
									{
										Node: gql.Node{
											FilterOnly: "false",
											Label:      "Region",
											Name:       "Region",
										},
									},
								},
							},
						},
					},
				},
				{
					Node: gql.Node{
						Name:       "Health",
						Label:      "Health",
						Categories: gql.Categories{TotalCount: 6},
						MapFrom:    []gql.Variables{},
					},
				},
				{
					Node: gql.Node{
						Name:       "Marital Status",
						Label:      "Marital Status",
						Categories: gql.Categories{TotalCount: 5},
						MapFrom:    []gql.Variables{},
					},
				},
				{
					Node: gql.Node{
						Name:       "Region",
						Label:      "Region",
						Categories: gql.Categories{TotalCount: 10},
						MapFrom:    []gql.Variables{},
					},
				},
				{
					Node: gql.Node{
						Name:       "Sex",
						Label:      "Sex",
						Categories: gql.Categories{TotalCount: 2},
						MapFrom:    []gql.Variables{},
					},
				},
			},
		},
	},
}

// mockRespBodyGetGeographyDimensions is a successful 'get geography dimensions' query respose that is returned from a mocked client for testing
var mockRespBodyGetGeographyDimensions = `
{
	"data": {
		"dataset": {
			"ruleBase": {
				"isSourceOf": {
					"totalCount": 2,
					"edges": [
						{
							"node": {
								"categories": {
									"totalCount": 2
								},
								"label": "Country",
								"mapFrom": [
									{
										"edges": [
											{
												"node": {
													"filterOnly": "false",
													"label": "Region",
													"name": "Region"
												}
											}
										]
									}
								],
								"name": "Country"
							}
						},
						{
							"node": {
								"categories": {
									"totalCount": 10
								},
								"label": "Region",
								"mapFrom": [],
								"name": "Region"
							}
						}
					]
				},
				"name": "Region"
			}
		}
	}
}
`

var expectedGeographyDimensions = cantabular.GetGeographyDimensionsResponse{
	Dataset: gql.DatasetRuleBase{
		RuleBase: gql.RuleBase{
			IsSourceOf: gql.Variables{
				TotalCount: 2,
				Edges: []gql.Edge{
					{
						Node: gql.Node{
							Name:       "Country",
							Label:      "Country",
							Categories: gql.Categories{TotalCount: 2},
							MapFrom: []gql.Variables{
								{
									Edges: []gql.Edge{
										{
											Node: gql.Node{
												Name:       "Region",
												Label:      "Region",
												Categories: gql.Categories{TotalCount: 0},
												MapFrom:    []gql.Variables(nil),
												FilterOnly: "false",
											},
										},
									},
								},
							},
							FilterOnly: "",
						},
					},
					{
						Node: gql.Node{
							Name:       "Region",
							Label:      "Region",
							Categories: gql.Categories{TotalCount: 10},
							MapFrom:    []gql.Variables{},
							FilterOnly: "",
						},
					},
				},
			},
			Name: "Region",
		},
	},
	PaginationResponse: cantabular.PaginationResponse{
		Count:      2,
		TotalCount: 2,
		PaginationParams: cantabular.PaginationParams{
			Limit:  10,
			Offset: 0,
		},
	},
}

// mockRespBodyGetDimensionsByName is a successful 'get dimensions by name' query respose that is returned from a mocked client for testing
var mockRespBodyGetDimensionsByName = `{
	"data": {
		"dataset": {
			"variables": {
				"edges": [
					{
						"node": {
							"categories": {
								"totalCount": 8
							},
							"label": "Age",
							"mapFrom": [],
							"name": "Age"
						}
					},
					{
						"node": {
							"categories": {
								"totalCount": 10
							},
							"label": "Region",
							"mapFrom": [],
							"name": "Region"
						}
					}
				]
			}
		}
	}
}`

// expectedDimensionsByName is the expected response struct generated from a successful 'get dimensions by name' query for testing
var expectedDimensionsByName = cantabular.GetDimensionsResponse{
	Dataset: gql.DatasetVariables{
		Variables: gql.Variables{
			Edges: []gql.Edge{
				{
					Node: gql.Node{
						Name:       "Age",
						Label:      "Age",
						Categories: gql.Categories{TotalCount: 8},
						MapFrom:    []gql.Variables{},
					},
				},
				{
					Node: gql.Node{
						Name:       "Region",
						Label:      "Region",
						Categories: gql.Categories{TotalCount: 10},
						MapFrom:    []gql.Variables{},
					},
				},
			},
		},
	},
}

// mockRespBodySearchDimensions is a successful 'search dimensions' query respose that is returned from a mocked client for testing
var mockRespBodySearchDimensions = `{
	"data": {
		"dataset": {
			"variables": {
				"search": {
					"edges": [
						{
							"node": {
								"label": "Country",
								"mapFrom": [
									{
										"edges": [
											{
												"node": {
													"label": "Region",
													"name": "Region"
												}
											}
										],
										"totalCount": 1
									}
								],
								"name": "Country"
							}
						},
						{
							"node": {
								"label": "Country of Birth",
								"mapFrom": [],
								"name": "Country of Birth"
							}
						}
					]
				}
			}
		}
	}
}
`

// expectedSearchDimensionsResponse is the expected response struct generated from a successful 'search dimensions' query for testing
var expectedSearchDimensionsResponse = cantabular.GetDimensionsResponse{
	Dataset: gql.DatasetVariables{
		Variables: gql.Variables{
			Search: gql.Search{
				Edges: []gql.Edge{
					{
						Node: gql.Node{
							Name:  "Country",
							Label: "Country",
							MapFrom: []gql.Variables{
								{
									TotalCount: 1,
									Edges: []gql.Edge{
										{
											Node: gql.Node{
												Name:  "Region",
												Label: "Region",
											},
										},
									},
								},
							},
						},
					},
					{
						Node: gql.Node{
							Name:    "Country of Birth",
							Label:   "Country of Birth",
							MapFrom: []gql.Variables{},
						},
					},
				},
			},
		},
	},
}

// mockRespBodyGetDimensionOptions is a successful 'get dimension options' query respose that is returned from a mocked client for testing
var mockRespBodyGetDimensionOptions = `
{
    "data": {
        "dataset": {
            "table": {
                "dimensions": [
                    {
                        "categories": [
                            {
                                "code": "E",
                                "label": "England"
                            },
                            {
                                "code": "W",
                                "label": "Wales"
                            }
                        ],
                        "variable": {
                            "label": "Country",
                            "name": "Country"
                        }
                    },
                    {
                        "categories": [
                            {
                                "code": "1",
                                "label": "0 to 15"
                            },
                            {
                                "code": "2",
                                "label": "16 to 24"
                            },
                            {
                                "code": "3",
                                "label": "25 to 34"
                            },
                            {
                                "code": "4",
                                "label": "35 to 44"
                            },
                            {
                                "code": "5",
                                "label": "45 to 54"
                            },
                            {
                                "code": "6",
                                "label": "55 to 64"
                            },
                            {
                                "code": "7",
                                "label": "65 to 74"
                            },
                            {
                                "code": "8",
                                "label": "75 and over"
                            }
                        ],
                        "variable": {
                            "label": "Age",
                            "name": "Age"
                        }
                    },
                    {
                        "categories": [
                            {
                                "code": "1",
                                "label": "Managers, Directors and Senior Officials"
                            },
                            {
                                "code": "2",
                                "label": "Professional Occupations"
                            },
                            {
                                "code": "3",
                                "label": "Associate Professional and Technical Occupations"
                            },
                            {
                                "code": "4",
                                "label": "Administrative and Secretarial Occupations"
                            },
                            {
                                "code": "5",
                                "label": "Skilled Trades Occupations"
                            },
                            {
                                "code": "6",
                                "label": "Caring, Leisure and Other Service Occupations"
                            },
                            {
                                "code": "7",
                                "label": "Sales and Customer Service Occupations"
                            },
                            {
                                "code": "8",
                                "label": "Process, Plant and Machine Operatives"
                            },
                            {
                                "code": "9",
                                "label": "Elementary Occupations"
                            },
                            {
                                "code": "-9",
                                "label": "N/A"
                            }
                        ],
                        "variable": {
                            "label": "Occupation",
                            "name": "Occupation"
                        }
                    }
                ]
            }
        }
    }
}`

// expectedDimensionOptions is the expected response struct generated from a successful 'get dimension options' query for testing
var expectedDimensionOptions = cantabular.GetDimensionOptionsResponse{
	Dataset: cantabular.StaticDatasetDimensionOptions{
		Table: cantabular.DimensionsTable{
			Dimensions: []cantabular.Dimension{
				{
					Variable: cantabular.VariableBase{
						Name:  "Country",
						Label: "Country",
					},
					Categories: []cantabular.Category{
						{Code: "E", Label: "England"},
						{Code: "W", Label: "Wales"},
					},
				},
				{
					Variable: cantabular.VariableBase{
						Label: "Age",
						Name:  "Age",
					},
					Categories: []cantabular.Category{
						{Code: "1", Label: "0 to 15"},
						{Code: "2", Label: "16 to 24"},
						{Code: "3", Label: "25 to 34"},
						{Code: "4", Label: "35 to 44"},
						{Code: "5", Label: "45 to 54"},
						{Code: "6", Label: "55 to 64"},
						{Code: "7", Label: "65 to 74"},
						{Code: "8", Label: "75 and over"},
					},
				},
				{
					Variable: cantabular.VariableBase{
						Name:  "Occupation",
						Label: "Occupation",
					},
					Categories: []cantabular.Category{
						{Code: "1", Label: "Managers, Directors and Senior Officials"},
						{Code: "2", Label: "Professional Occupations"},
						{Code: "3", Label: "Associate Professional and Technical Occupations"},
						{Code: "4", Label: "Administrative and Secretarial Occupations"},
						{Code: "5", Label: "Skilled Trades Occupations"},
						{Code: "6", Label: "Caring, Leisure and Other Service Occupations"},
						{Code: "7", Label: "Sales and Customer Service Occupations"},
						{Code: "8", Label: "Process, Plant and Machine Operatives"},
						{Code: "9", Label: "Elementary Occupations"},
						{Code: "-9", Label: "N/A"},
					},
				},
			},
		},
	},
}

var mockRespBodyGetAreas = `
{
  "data": {
    "dataset": {
      "ruleBase": {
        "isSourceOf": {
          "search": {
            "edges": [
              {
                "node": {
                  "label": "City",
                  "name": "city",
                  "categories": {
                    "search": {
                      "edges": [
                        {
                          "node": {
                            "code": "0",
                            "label": "London"
                          }
                        }
                      ]
                    }
                  }
                }
              }
            ]
          }
        }
      }
    }
  }
}
`

var expectedAreas = cantabular.GetAreasResponse{
	Dataset: gql.DatasetRuleBase{
		RuleBase: gql.RuleBase{
			IsSourceOf: gql.Variables{
				Search: gql.Search{
					Edges: []gql.Edge{
						{
							Node: gql.Node{
								Name:  "city",
								Label: "City",
								Categories: gql.Categories{
									Search: gql.Search{
										Edges: []gql.Edge{
											{
												Node: gql.Node{
													Code:  "0",
													Label: "London",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
