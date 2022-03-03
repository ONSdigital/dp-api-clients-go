package interactives

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/pkg/errors"
)

const (
	userAuthToken    = "amatoken"
	serviceAuthToken = "amaservicetoken"
	testHost         = "http://localhost:8080"
)

var (
	ctx                = context.Background()
	successful, failed = true, false
	checkRequestBase   = func(httpClient *dphttp.ClienterMock, expectedMethod string, expectedUri string) {
		So(len(httpClient.DoCalls()), ShouldEqual, 1)
		So(httpClient.DoCalls()[0].Req.URL.RequestURI(), ShouldResemble, expectedUri)
		So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
		So(httpClient.DoCalls()[0].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+serviceAuthToken)
	}
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

func TestClient_GetInteractives(t *testing.T) {
	offset := 1
	limit := 10

	Convey("given a 200 status is returned", t, func() {
		expectedInteractives := List{
			Items: []Interactive{
				{ID: "InteractivesID1"},
				{ID: "InteractivesID2"},
			},
			Count:      2,
			Offset:     offset,
			Limit:      limit,
			TotalCount: 3,
		}
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, expectedInteractives, nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when GetInteractives is called with valid values for limit and offset", func() {
			q := QueryParams{Offset: offset, Limit: limit}
			actualInteractives, err := interactivesClient.GetInteractives(ctx, userAuthToken, serviceAuthToken, &q)

			Convey("a positive response is returned, with the expected interactives", func() {
				So(err, ShouldBeNil)
				So(actualInteractives, ShouldResemble, expectedInteractives)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := fmt.Sprintf("/interactives?offset=%d&limit=%d", offset, limit)
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})

		Convey("when GetInteractives is called with negative offset", func() {
			q := QueryParams{Offset: -1, Limit: limit}
			options, err := interactivesClient.GetInteractives(ctx, userAuthToken, serviceAuthToken, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, List{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("when GetInteractives is called with negative limit", func() {
			q := QueryParams{Offset: offset, Limit: -1}
			options, err := interactivesClient.GetInteractives(ctx, userAuthToken, serviceAuthToken, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, List{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, List{}, nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when GetInteractives is called", func() {
			options, err := interactivesClient.GetInteractives(ctx, userAuthToken, serviceAuthToken, nil)

			Convey("the expected error response is returned, with an empty options struct", func() {
				So(err, ShouldResemble, &ErrInvalidInteractivesAPIResponse{
					actualCode: 404,
					uri:        "http://localhost:8080/interactives",
					body:       "{\"items\":null,\"count\":0,\"offset\":0,\"limit\":0,\"total_count\":0}",
				})
				So(options, ShouldResemble, List{})
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := "/interactives"
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})
	})
}

func TestClient_PutInteractive(t *testing.T) {
	checkResponse := func(httpClient *dphttp.ClienterMock, expectedInteractive InteractiveUpdate) {

		checkRequestBase(httpClient, http.MethodPut, "/interactives/123")

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)

		var actualInteractive InteractiveUpdate
		err := json.Unmarshal(actualBody, &actualInteractive)
		So(err, ShouldBeNil)
		So(actualInteractive, ShouldResemble, expectedInteractive)
	}

	Convey("Given a valid interactive", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, "", nil})
		interactiveClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			inter := InteractiveUpdate{ImportSuccessful: &successful}
			err := interactiveClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", inter)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, inter)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular error")
		httpClient := createHTTPClientMockErr(mockErr)
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			v := InteractiveUpdate{ImportSuccessful: &failed}
			err := interactivesClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusInternalServerError, "", nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			v := InteractiveUpdate{ImportSuccessful: &failed}
			err := interactivesClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 500 from interactives api: http://localhost:8080/interactives/123, body: ").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})
		})
	})
}

func newInteractivesClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, httpClient)
	interactivesClient := NewWithHealthClient(healthClient)
	return interactivesClient
}

func createHTTPClientMockErr(err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return nil, err
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func createHTTPClientMock(mockedHTTPResponse ...MockedHTTPResponse) *dphttp.ClienterMock {
	numCall := 0
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := json.Marshal(mockedHTTPResponse[numCall].Body)
			resp := &http.Response{
				StatusCode: mockedHTTPResponse[numCall].StatusCode,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
				Header:     http.Header{},
			}
			for hKey, hVal := range mockedHTTPResponse[numCall].Headers {
				resp.Header.Set(hKey, hVal)
			}
			numCall++
			return resp, nil
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}
