package cantabular_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/mock"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

func TestStream(t *testing.T) {
	Convey("Given a stream consumer that scans an io.Reader", t, func() {
		out := ""
		consume := func(ctx context.Context, r io.Reader) error {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				out += fmt.Sprintln(line)
			}
			return nil
		}

		Convey("And an http client that returns a valid query response and 200 OK status code", func() {
			mockHttpClient := &dphttp.ClienterMock{
				PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
					return Response(
						[]byte(mockRespBodyStaticDataset),
						http.StatusOK,
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

			Convey("Then the expected CSV is successfully streamed with the expected number of rows", func() {
				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "Example",
					Variables: []string{"city", "siblings"},
				}
				rowCount, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
				So(err, ShouldBeNil)
				So(out, ShouldResemble, expectedCsv)
				So(rowCount, ShouldEqual, 22)
			})

			Convey("Then calling stream with a cancelled context results in the expected error being returned and only the first line being processed", func() {
				testCtxWithCancel, cancel := context.WithCancel(testCtx)
				cancel()

				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "Example",
					Variables: []string{"city", "siblings"},
				}
				_, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtxWithCancel, req, consume)
				So(err, ShouldResemble,
					fmt.Errorf("transform error: %w",
						fmt.Errorf("error decoding table fields: %w",
							fmt.Errorf("error decoding values: %w",
								fmt.Errorf("error iterating to next row: %w",
									fmt.Errorf("context is done: %w", errors.New("context canceled")))))))
				So(out, ShouldEqual, "count,City,Number of siblings\n1,London,No siblings\n")
			})
		})

		Convey("And an http client that returns a 'dataset not loaded in this server' error response and 200 OK status code, due to a wrong 'dataset' value in the query", func() {
			mockHttpClient := &dphttp.ClienterMock{
				PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
					return Response(
						[]byte(mockRespBodyNoDataset),
						http.StatusOK,
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

			Convey("Then the expected error is returned and nothing is streamed", func() {
				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "InexistentDataset",
					Variables: []string{"city", "siblings"},
				}
				_, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
				So(err, ShouldResemble, fmt.Errorf("transform error: %w",
					fmt.Errorf("error(s) returned by graphQL query: %w",
						errors.New("404 Not Found: dataset not loaded in this server"))))
				So(out, ShouldEqual, "")
			})
		})

		Convey("And an http client that returns a 'variable at position 1' error response and 200 OK status code, due to a wrong variable value in the query", func() {
			mockHttpClient := &dphttp.ClienterMock{
				PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
					return Response(
						[]byte(mockRespBodyNoTable),
						http.StatusOK,
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

			Convey("Then the expected error is returned and nothing is streamed", func() {
				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "Example",
					Variables: []string{"wrong", "siblings"},
				}
				_, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
				So(err, ShouldResemble, fmt.Errorf("transform error: %w",
					fmt.Errorf("error(s) returned by graphQL query: %w",
						errors.New("400 Bad Request: variable at position 1 does not exist"))))
				So(out, ShouldEqual, "")
			})
		})

		Convey("And an http client that refuses to connect", func() {
			mockHttpClient := &dphttp.ClienterMock{
				PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
					return nil, errors.New(`Post "cantabular.ext.host/graphql": dial tcp 127.0.0.1:8493: connect: connection refused`)
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

			Convey("Then the expected error is returned and nothing is streamed", func() {
				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "Example",
					Variables: []string{"city", "siblings"},
				}
				_, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
				So(err.Error(), ShouldResemble, `failed to make GraphQL query: failed to make request: Post "cantabular.ext.host/graphql": dial tcp 127.0.0.1:8493: connect: connection refused`)
				So(out, ShouldEqual, "")
			})
		})

		Convey("And an http client that returns status code 503", func() {
			mockHttpClient := &dphttp.ClienterMock{
				PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
					return Response(
						[]byte(`{"message": "something is broken"}`),
						http.StatusBadGateway,
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

			Convey("Then the expected error is returned and nothing is streamed", func() {
				req := cantabular.StaticDatasetQueryRequest{
					Dataset:   "Example",
					Variables: []string{"city", "siblings"},
				}
				_, err := cantabularClient.StaticDatasetQueryStreamCSV(testCtx, req, consume)
				So(err, ShouldResemble, dperrors.New(
					errors.New("something is broken"),
					http.StatusBadGateway,
					log.Data{
						"url": "cantabular.ext.host/graphql",
					},
				))
				So(out, ShouldEqual, "")
			})
		})
	})
}

func TestStaticDatasetQueryHappy(t *testing.T) {

	Convey("Given a correct response from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}
		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				return nil
			},
		}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host:       "cantabular.host",
				ExtApiHost: "cantabular.ext.host",
			},
			mockHttpClient,
			mockGQLClient,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestStaticDatasetQueryUnHappy(t *testing.T) {

	Convey("Given the graphQL Client is not configured", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host: "cantabular.host",
			},
			mockHttpClient,
			nil,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
			So(err, ShouldNotBeNil)

			Convey("Status Code 503 Service Unavailable should be recoverable from error", func() {
				_, err := cantabularClient.StaticDatasetQuery(testCtx, req)
				So(dperrors.StatusCode(err), ShouldEqual, http.StatusServiceUnavailable)
			})
		})
	})

	Convey("Given a GraphQL error from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{}
		mockGQLClient := &mock.GraphQLClientMock{
			QueryFunc: func(ctx context.Context, query interface{}, vars map[string]interface{}) error {
				if q, ok := query.(*cantabular.StaticDatasetQuery); ok {
					q.Dataset.Table.Error = "I am error response"
					return nil
				}
				return errors.New("query could not be cast to correct type")
			},
		}

		cantabularClient := cantabular.NewClient(
			cantabular.Config{
				Host:       "cantabular.host",
				ExtApiHost: "cantabular.ext.host",
			},
			mockHttpClient,
			mockGQLClient,
		)

		Convey("When the StaticDatasetQuery method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{}
			_, err := cantabularClient.StaticDatasetQuery(testCtx, req)

			Convey("An error should be returned with status code 400 Bad Request", func() {
				So(err, ShouldNotBeNil)
				So(dperrors.StatusCode(err), ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}

func TestGetDimensionsHappy(t *testing.T) {
	Convey("Given a correct getDimensions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyGetDimensions),
					http.StatusOK,
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

		Convey("When the GetDimensions method is called", func() {
			resp, err := cantabularClient.GetDimensions(testCtx, "Teaching-Dataset")

			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensions)
			})
		})
	})
}

func TestGetDimensionsUnhappy(t *testing.T) {
	Convey("Given a mocked cantabular client that returns a no-dataset graphql error", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyNoDataset),
					http.StatusOK,
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

		Convey("Then the expected error is returned", func() {
			resp, err := cantabularClient.GetDimensions(testCtx, "InexistentDataset")

			So(err, ShouldResemble, dperrors.New(
				errors.New("error(s) returned by graphQL query"),
				http.StatusOK,
				log.Data{
					"errors": []gql.Error{
						{
							Message:   "404 Not Found: dataset not loaded in this server",
							Path:      []string{"dataset"},
							Locations: []gql.Location{{Line: 2, Column: 2}},
						},
					},
				},
			))
			So(resp, ShouldBeNil)
		})
	})
}

func TestGetDimensionsByNameHappy(t *testing.T) {
	Convey("Given a correct getDimensions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyGetDimensionsByName),
					http.StatusOK,
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

		Convey("When the GetDimensions method is called", func() {
			resp, err := cantabularClient.GetDimensionsByName(testCtx, cantabular.StaticDatasetQueryRequest{
				Dataset:   "Teaching-Dataset",
				Variables: []string{"Age", "Region"},
			})

			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensionsByName)
			})
		})
	})
}

func TestGetDimensionOptionsHappy(t *testing.T) {
	Convey("Given a correct getDimensionOptions response from the /graphql endpoint", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyGetDimensionOptions),
					http.StatusOK,
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

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{
				Dataset:   "Teaching-Dataset",
				Variables: []string{"Country", "Age", "Occupation"},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("No error should be returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("And the expected response is returned", func() {
				So(*resp, ShouldResemble, expectedDimensionOptions)
			})
		})
	})
}

func TestGetDimensionOptionsUnhappy(t *testing.T) {
	testCtx := context.Background()

	Convey("Given a mocked cantabular client that returns a no-dataset graphql error", t, func() {
		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyNoDataset),
					http.StatusOK,
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

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{
				Dataset:   "InexistentDataset",
				Variables: []string{"Country", "Age", "Occupation"},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New("error(s) returned by graphQL query"),
					http.StatusOK,
					log.Data{
						"errors": []gql.Error{
							{
								Message:   "404 Not Found: dataset not loaded in this server",
								Path:      []string{"dataset"},
								Locations: []gql.Location{{Line: 2, Column: 2}},
							},
						},
					},
				))
				So(resp, ShouldBeNil)
			})
		})
	})

	Convey("Given a mocked cantabular client that returns a no-variable graphql error", t, func() {
		mockHttpClient := &dphttp.ClienterMock{
			PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
				return Response(
					[]byte(mockRespBodyNoVariable),
					http.StatusOK,
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

		Convey("When the GetDimensionOptions method is called", func() {
			req := cantabular.StaticDatasetQueryRequest{
				Dataset:   "Teaching-Dataset",
				Variables: []string{"Country", "Age", "inexistentVariable"},
			}
			resp, err := cantabularClient.GetDimensionOptions(testCtx, req)

			Convey("Then the expected error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New("error(s) returned by graphQL query"),
					http.StatusOK,
					log.Data{
						"errors": []gql.Error{
							{
								Message:   "400 Bad Request: variable at position 3 does not exist",
								Path:      []string{"dataset", "table"},
								Locations: []gql.Location{{Line: 4, Column: 3}},
							},
						},
					},
				))
				So(resp, ShouldBeNil)
			})
		})
	})
}

// mockRespBodyStaticDataset is a successful static dataset query respose that is returned from a mocked client for testing
var mockRespBodyStaticDataset = `
{
	"data": {
		"dataset": {
			"table": {
				"dimensions": [
					{
						"categories": [
							{"code": "0", "label": "London"},
							{"code": "1", "label": "Liverpool"},
							{"code": "2", "label": "Belfast"}
						],
						"count": 3,
						"variable": {"label": "City", "name": "city"}
					},
					{
						"categories": [
							{"code": "0", "label": "No siblings"},
							{"code": "1", "label": "1 sibling"},
							{"code": "2", "label": "2 siblings"},
							{"code": "3", "label": "3 siblings"},
							{"code": "4", "label": "4 siblings"},
							{"code": "5", "label": "5 siblings"},
							{"code": "6", "label": "6 or more siblings"}
						],
						"count":7,
						"variable": {"label": "Number of siblings", "name": "siblings"}
					}
				],
				"error": null,
				"values": [1,0,0,1,0,0,0,0,0,0,0,1,0,0,0,0,1,0,0,1,1]
			}
		}
	}
}`

// expectedCsv is the expected CSV generated from a successful static dataset query for testing
var expectedCsv = `count,City,Number of siblings
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

// mockRespBodyNoDataset is an error response that is returned from a mocked client for testing
// when a wrong Dataset is provided in the query
var mockRespBodyNoDataset = `
{
	"data": {
		"dataset": null
	},
	"errors": [
		{
			"message": "404 Not Found: dataset not loaded in this server",
			"locations": [
				{
					"line": 2,
					"column": 2
				}
			],
			"path": [
				"dataset"
			]
		}
	]
}`

// mockRespBodyNoVariable is an error response that is returned from a mocked client for testing
// when a wrong Variable name is provided in the query
var mockRespBodyNoVariable = `
{
	"data": {
		"dataset": null
	},
	"errors": [
		{
			"message": "400 Bad Request: variable at position 3 does not exist",
			"locations": [
				{
					"line": 4,
					"column": 3
				}
			],
			"path": [
				"dataset",
				"table"
			]
		}
	]
}`

// mockRespBodyNoTable is an error response that is returned from a mocked client for testing
// when a wrong variable is provided in the query
var mockRespBodyNoTable = `
{
	"data": {
		"dataset": {
			"table": null
		}
	},
	"errors": [
		{
			"message": "400 Bad Request: variable at position 1 does not exist",
			"locations": [
				{
					"line": 3,
					"column": 3
				}
			],
			"path": [
				"dataset",
				"table"
			]
		}
	]
}`
