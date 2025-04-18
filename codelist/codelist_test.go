package codelist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testServiceAuthToken = "666"
	testUserAuthToken    = "217" // room 217 the overlook hotel
	testHost             = "http://localhost:8080"
)

var initialState = health.CreateCheckState(service)

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{path, "/healthcheck"}
			},
		}

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)
		check := initialState

		Convey("when codelistClient.Checker is called", func() {
			err := codelistClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{path, "/healthcheck"}
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
				}, nil
			},
		}

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)
		check := initialState

		Convey("when codelistClient.Checker is called", func() {
			err := codelistClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{path, "/healthcheck"}
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, nil
			},
		}

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)
		check := initialState

		Convey("when codelistClient.Checker is called", func() {
			err := codelistClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 404)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{path, "/healthcheck"}
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 429,
				}, nil
			},
		}

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)
		check := initialState

		Convey("when codelistClient.Checker is called", func() {
			err := codelistClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, 429)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{path, "/healthcheck"}
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			},
		}

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)
		check := initialState

		Convey("when codelistClient.Checker is called", func() {
			err := codelistClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure(), ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_GetValues(t *testing.T) {
	host := "localhost:8080"

	Convey("should return expect values for 200 status response", t, func() {
		b, err := json.Marshal(testDimensionValues)
		So(err, ShouldBeNil)

		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		actual, err := codelistClient.GetValues(nil, testUserAuthToken, testServiceAuthToken, "999")

		So(err, ShouldBeNil)
		So(actual, ShouldResemble, testDimensionValues)

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		assertClienterDoCalls(req, "/code-lists/999/codes", host)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expect error if clienter.Do returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		actual, err := codelistClient.GetValues(nil, testUserAuthToken, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(actual, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		assertClienterDoCalls(req, "/code-lists/999/codes", host)
	})

	Convey("should return expected error for non 200 response status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 500)

		clienter := getClienterMock(resp, nil)
		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		expectedURI := fmt.Sprintf("%s/code-lists/%s/codes", testHost, "999")
		expectedErr := &ErrInvalidCodelistAPIResponse{http.StatusOK, 500, expectedURI}

		dimensionValues, err := codelistClient.GetValues(nil, testUserAuthToken, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		assertClienterDoCalls(req, "/code-lists/999/codes", host)
		So(body.IsClosed, ShouldBeTrue)
	})

	Convey("should return expected error if ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("lets get schwifty")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)

		clienter := getClienterMock(resp, nil)
		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		dimensionValues, err := codelistClient.GetValues(nil, testUserAuthToken, testServiceAuthToken, "999")

		So(err, ShouldResemble, expectedErr)
		So(dimensionValues, ShouldResemble, DimensionValues{})

		calls := clienter.DoCalls()
		So(calls, ShouldHaveLength, 1)

		req := calls[0].Req
		assertClienterDoCalls(req, "/code-lists/999/codes", host)
		So(body.IsClosed, ShouldBeTrue)
	})
}

func TestClient_GetIDNameMap(t *testing.T) {

	uri := "/code-lists/666/codes"
	host := "localhost:8080"

	Convey("give client.Do returns an error", t, func() {
		expectedErr := errors.New("bork")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistClient.GetIDNameMap(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		expectedErr := &ErrInvalidCodelistAPIResponse{
			expectedCode: http.StatusOK,
			actualCode:   403,
			uri:          testHost + uri,
		}

		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 403)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistClient.GetIDNameMap(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("i wander out where you can't see inside my shell i wait and bleed")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistClient.GetIDNameMap(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldEqual, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given unmarshalling the response body returns error", t, func() {
		// return bytes incompatible with the expected return type
		b := httpmocks.GetEntityBytes(t, []int{1, 2, 3, 4, 5, 6})
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistClient.GetIDNameMap(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given a successful http response is returned", t, func() {
		b := httpmocks.GetEntityBytes(t, testDimensionValues)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetIDNameMap is called", func() {
			actual, err := codelistClient.GetIDNameMap(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected ID Name map is returned", func() {
				expected := map[string]string{"123": "Schwifty"}
				So(actual, ShouldResemble, expected)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetGeographyCodeLists(t *testing.T) {
	uri := "/code-lists"
	host := "localhost:8080"
	query := "type=geography"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("master master obey your master")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistClient.GetGeographyCodeLists(nil, testUserAuthToken, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, 500)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistClient.GetGeographyCodeLists(nil, testUserAuthToken, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})

				expectedErr := &ErrInvalidCodelistAPIResponse{
					expectedCode: http.StatusOK,
					actualCode:   500,
					uri:          testHost + uri + "?" + query,
				}

				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("peace sells but who buying")
		body := httpmocks.NewReadCloserMock([]byte{}, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistClient.GetGeographyCodeLists(nil, testUserAuthToken, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarshal returns an error", t, func() {
		entity := []int{1, 666, 8, 16}
		b := httpmocks.GetEntityBytes(t, entity) // return bytes that are incompatible with the expected return type
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistClient.GetGeographyCodeLists(nil, testUserAuthToken, testServiceAuthToken)

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeListResults{})
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistClient is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, testCodeListResults)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistClient.GetGeographyCodeLists is called", func() {
			actual, err := codelistClient.GetGeographyCodeLists(nil, testUserAuthToken, testServiceAuthToken)

			Convey("then the expected result is returned", func() {
				So(actual, ShouldResemble, testCodeListResults)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetCodeListEditions(t *testing.T) {
	uri := "/code-lists/666/editions"
	host := "localhost:8080"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("smashing through the boundaries lunacy has found me cannot stop the battery")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeListEditions is called", func() {
			actual, err := codelistClient.GetCodeListEditions(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, EditionsListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		expectedErr := &ErrInvalidCodelistAPIResponse{
			expectedCode: http.StatusOK,
			actualCode:   http.StatusBadRequest,
			uri:          "http://" + host + uri,
		}

		body := httpmocks.NewReadCloserMock(nil, nil)
		resp := httpmocks.NewResponseMock(body, 400)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeListEditions is called", func() {
			actual, err := codelistClient.GetCodeListEditions(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, EditionsListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("have you run your fingers down the wall and have you felt your neck skin crawl when youre searching for the light")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeListEditions is called", func() {
			actual, err := codelistClient.GetCodeListEditions(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the expected error is returned", func() {
				So(actual, ShouldResemble, EditionsListResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarshal returns an error", t, func() {
		i := 666
		b := httpmocks.GetEntityBytes(t, i) // return a value that cannot be marshalled into the expected struct
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeListEditions is called", func() {
			actual, err := codelistClient.GetCodeListEditions(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("then client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the expected error is returned", func() {
				So(actual, ShouldResemble, EditionsListResults{})
				So(err, ShouldNotBeNil)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistclient.GetCodeListEditions is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, editionsListResults)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, 200)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeListEditions is called", func() {
			actual, err := codelistClient.GetCodeListEditions(nil, testUserAuthToken, testServiceAuthToken, "666")

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("then the expected value is returned", func() {
				So(actual, ShouldResemble, editionsListResults)
				So(err, ShouldBeNil)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetCodes(t *testing.T) {
	uri := "/code-lists/foo/editions/bar/codes"
	host := "localhost:8080"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("generals gathered in their masses, just like witches at black masses")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodes is called", func() {
			actual, err := codelistClient.GetCodes(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodesResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 status", t, func() {
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodes is called", func() {
			actual, err := codelistClient.GetCodes(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodesResults{})
				So(err, ShouldResemble, &ErrInvalidCodelistAPIResponse{
					http.StatusOK,
					http.StatusInternalServerError,
					"http://" + host + uri,
				})
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("exit light enter night take my hand we're off to never-never land")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodes is called", func() {
			actual, err := codelistClient.GetCodes(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodesResults{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarshal returns an error", t, func() {
		v := []int{0}
		b := httpmocks.GetEntityBytes(t, v) // return bytes that cannot be marshalled into the expected struct
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodes is called", func() {
			actual, err := codelistClient.GetCodes(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodesResults{})
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistclient.GetCodes is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, codesResults)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodes is called", func() {
			actual, err := codelistClient.GetCodes(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, codesResults)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetCodeByID(t *testing.T) {
	uri := "/code-lists/foo/editions/bar/codes/1"
	host := "localhost:8080"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("quoth the raven 'nevermore'")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetCodeByID(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 status response", t, func() {
		expectedErr := &ErrInvalidCodelistAPIResponse{
			http.StatusOK,
			http.StatusInternalServerError,
			"http://" + host + uri,
		}

		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetCodeByID(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("i know what you're thinking did he fire six shots or only five")

		body := httpmocks.NewReadCloserMock([]byte{}, expectedErr)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetCodeByID(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarshal returns an error", t, func() {
		v := []int{0}
		b := httpmocks.GetEntityBytes(t, v) // return bytes that cannot be marshalled into the expected struct
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetCodeByID(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, CodeResult{})
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistclient.GetCodeByID is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, codeResult)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetCodeByID(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, codeResult)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestClient_GetDatasetsByCode(t *testing.T) {
	uri := "/code-lists/foo/editions/bar/codes/1/datasets"
	host := "localhost:8080"

	Convey("given clienter.Do returns an error", t, func() {
		expectedErr := errors.New("murders in the rue morgue")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetDatasetsByCode(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, DatasetsResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})
		})
	})

	Convey("given clienter.Do returns a non 200 response status", t, func() {
		expectedErr := &ErrInvalidCodelistAPIResponse{
			http.StatusOK,
			http.StatusInternalServerError,
			"http://" + host + uri,
		}

		body := httpmocks.NewReadCloserMock(make([]byte, 0), nil)
		resp := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetDatasetsByCode(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, DatasetsResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given ioutil.ReadAll returns an error", t, func() {
		expectedErr := errors.New("error mcerrorface")
		body := httpmocks.NewReadCloserMock(make([]byte, 0), expectedErr)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetDatasetsByCode(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, DatasetsResult{})
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given json.Unmarhal returns an error", t, func() {
		v := "golang or go home"
		b := httpmocks.GetEntityBytes(t, v)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetDatasetsByCode(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, DatasetsResult{})
				So(err, ShouldNotBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("given codelistclient.GetDatasetsByCode is successful", t, func() {
		b := httpmocks.GetEntityBytes(t, datasetsResult)
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		clienter := getClienterMock(resp, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when codelistclient.GetCodeByID is called", func() {
			actual, err := codelistClient.GetDatasetsByCode(nil, testUserAuthToken, testServiceAuthToken, "foo", "bar", "1")

			Convey("then the expected error is returned", func() {
				So(actual, ShouldResemble, datasetsResult)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called 1 time with the expected parameters", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				assertClienterDoCalls(req, uri, host)
			})

			Convey("and the response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

}

func TestDoGetWithAuthHeaders(t *testing.T) {
	Convey("given create new request returns an error", t, func() {
		clienter := getClienterMock(nil, nil)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when doGetWithAuthHeaders is called", func() {
			resp, err := codelistClient.doGetWithAuthHeaders(nil, testUserAuthToken, testServiceAuthToken, "@£$%^&*()_")

			Convey("then the expected error is returned", func() {
				So(resp, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("and clienter is not called", func() {
				So(clienter.DoCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("given sending a request is unsuccessful", t, func() {
		expectedErr := errors.New("im going off the rails on a crazy train")
		clienter := getClienterMock(nil, expectedErr)

		hcCli := health.NewClientWithClienter("", testHost, clienter)
		codelistClient := NewWithHealthClient(hcCli)

		Convey("when doGetWithAuthHeaders is called", func() {
			resp, err := codelistClient.doGetWithAuthHeaders(nil, testUserAuthToken, testServiceAuthToken, "/foobar")

			Convey("then the expected error is returned", func() {
				So(resp, ShouldBeNil)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("and clienter is not called", func() {
				calls := clienter.DoCalls()
				So(calls, ShouldHaveLength, 1)

				req := calls[0].Req
				So(req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, dprequest.BearerPrefix+testServiceAuthToken)
				So(req.Header.Get(dprequest.FlorenceHeaderKey), ShouldEqual, testUserAuthToken)
			})
		})
	})
}

func TestSetAuthenticationHeaders(t *testing.T) {
	Convey("should return expected error if request is nil", t, func() {
		err := setAuthenticationHeaders(nil, testUserAuthToken, testServiceAuthToken)
		So(err, ShouldResemble, headers.ErrRequestNil)
	})

	Convey("should not return an error if user auth token is empty", t, func() {
		req := httptest.NewRequest(http.MethodGet, testHost, nil)
		err := setAuthenticationHeaders(req, "", "")
		So(err, ShouldBeNil)

		actual, getErr := headers.GetUserAuthToken(req)
		So(actual, ShouldBeEmpty)
		So(getErr, ShouldResemble, headers.ErrHeaderNotFound)
	})

	Convey("should not return an error if service auth token is empty", t, func() {
		req := httptest.NewRequest(http.MethodGet, testHost, nil)
		err := setAuthenticationHeaders(req, testUserAuthToken, "")
		So(err, ShouldBeNil)

		actual, getErr := headers.GetUserAuthToken(req)
		So(actual, ShouldResemble, testUserAuthToken)
		So(getErr, ShouldBeNil)

	})

	Convey("should set expected user and service auth tokens", t, func() {
		req := httptest.NewRequest(http.MethodGet, testHost, nil)
		err := setAuthenticationHeaders(req, testUserAuthToken, testServiceAuthToken)
		So(err, ShouldBeNil)

		actual, getErr := headers.GetUserAuthToken(req)
		So(actual, ShouldResemble, testUserAuthToken)
		So(getErr, ShouldBeNil)

		actual, getErr = headers.GetServiceAuthToken(req)
		So(actual, ShouldResemble, testServiceAuthToken)
		So(getErr, ShouldBeNil)
	})
}

func getClienterMock(resp *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {
		},
	}
}

func assertClienterDoCalls(actual *http.Request, uri string, host string) {
	So(actual.URL.Path, ShouldEqual, uri)
	So(actual.URL.Host, ShouldEqual, host)
	So(actual.Method, ShouldEqual, "GET")
	So(actual.Body, ShouldBeNil)
	So(actual.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, dprequest.BearerPrefix+testServiceAuthToken)
	So(actual.Header.Get(dprequest.FlorenceHeaderKey), ShouldEqual, testUserAuthToken)
}
