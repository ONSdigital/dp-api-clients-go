package interactives

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/pkg/errors"
)

const (
	userAuthToken    = "amatoken"
	serviceAuthToken = "amaservicetoken"
	testHost         = "http://localhost:8080"
)

var (
	ctx              = context.Background()
	checkRequestBase = func(httpClient *dphttp.ClienterMock, expectedMethod string, expectedUri string) {
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

	Convey("given a 200 status is returned", t, func() {
		expectedInteractives := []Interactive{
			{ID: "InteractivesID1"},
			{ID: "InteractivesID2"},
		}
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, expectedInteractives, nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when GetInteractives is called with valid values for limit, offset and filter", func() {
			q := &InteractiveFilter{Metadata: &InteractiveMetadata{ResourceID: "resid123"}}
			i, err := interactivesClient.ListInteractives(ctx, userAuthToken, serviceAuthToken, q)

			Convey("a positive response is returned, with the expected interactives", func() {
				So(err, ShouldBeNil)
				So(i, ShouldResemble, expectedInteractives)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				marshal, _ := json.Marshal(q)
				expectedURI := fmt.Sprintf("/v1/interactives?filter=%s", url.QueryEscape(string(marshal)))
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, []Interactive{}, nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when GetInteractives is called", func() {
			i, err := interactivesClient.ListInteractives(ctx, userAuthToken, serviceAuthToken, nil)

			Convey("the expected error response is returned, with an empty options struct", func() {
				So(err, ShouldResemble, &ErrInvalidInteractivesAPIResponse{
					actualCode: 404,
					uri:        "http://localhost:8080/v1/interactives",
					body:       "[]",
				})
				So(i, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := "/v1/interactives"
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})
	})
}

func TestClient_PutInteractive(t *testing.T) {
	checkResponse := func(httpClient *dphttp.ClienterMock, expectedInteractive Interactive) {

		checkRequestBase(httpClient, http.MethodPut, "/v1/interactives/123")

		firstReq := httpClient.DoCalls()[0].Req
		err := firstReq.ParseMultipartForm(50)
		updateModelJson := firstReq.FormValue(UpdateFormFieldKey)

		var i Interactive
		err = json.Unmarshal([]byte(updateModelJson), &i)

		So(err, ShouldBeNil)
		So(i, ShouldResemble, expectedInteractive)
	}

	Convey("Given a valid interactive", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, "", nil})
		interactiveClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			i := Interactive{}
			err := interactiveClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", i)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, i)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular error")
		httpClient := createHTTPClientMockErr(mockErr)
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			i := Interactive{}
			err := interactivesClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", i)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, i)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusInternalServerError, "", nil})
		interactivesClient := newInteractivesClient(httpClient)

		Convey("when put interactive is called", func() {
			i := Interactive{}
			err := interactivesClient.PutInteractive(ctx, userAuthToken, serviceAuthToken, "123", i)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 500 from interactives api: http://localhost:8080/v1/interactives/123, body: ").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, i)
			})
		})
	})
}

func TestClient_PatchInteractive(t *testing.T) {
	checkResponse := func(httpClient *dphttp.ClienterMock, expected PatchRequest) {
		checkRequestBase(httpClient, http.MethodPatch, "/v1/interactives/123")

		firstReq := httpClient.DoCalls()[0].Req
		body, err := io.ReadAll(firstReq.Body)
		So(err, ShouldBeNil)

		var actual PatchRequest
		err = json.Unmarshal(body, &actual)
		So(err, ShouldBeNil)
		So(actual, ShouldResemble, expected)
	}

	Convey("Given a valid interactive", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{
			http.StatusOK,
			Interactive{},
			nil,
		})
		c := newInteractivesClient(httpClient)

		Convey("when valid patch request is called", func() {
			r := PatchRequest{
				Attribute:   PatchArchive,
				Interactive: Interactive{},
			}
			_, err := c.PatchInteractive(ctx, userAuthToken, serviceAuthToken, "123", r)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, r)
			})
		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular error")
		httpClient := createHTTPClientMockErr(mockErr)
		c := newInteractivesClient(httpClient)

		Convey("when valid patch request is called", func() {
			r := PatchRequest{
				Attribute:   PatchArchive,
				Interactive: Interactive{},
			}
			_, err := c.PatchInteractive(ctx, userAuthToken, serviceAuthToken, "123", r)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, r)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusInternalServerError, "", nil})
		c := newInteractivesClient(httpClient)

		Convey("when valid patch request is called", func() {
			r := PatchRequest{
				Attribute:   PatchArchive,
				Interactive: Interactive{},
			}
			_, err := c.PatchInteractive(ctx, userAuthToken, serviceAuthToken, "123", r)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 500 from interactives api: http://localhost:8080/v1/interactives/123, body: ").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, r)
			})
		})
	})
}

func TestClient_GetInterface(t *testing.T) {

	Convey("given a 200 status with valid empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{
			http.StatusOK,
			Interactive{},
			nil,
		})
		datasetClient := newInteractivesClient(httpClient)

		Convey("when GetInterface is called", func() {
			i, err := datasetClient.GetInteractive(ctx, userAuthToken, serviceAuthToken, "123")

			Convey("a positive response is returned with empty interface and the expected ETag", func() {
				So(err, ShouldBeNil)
				So(i, ShouldResemble, Interactive{})
			})

			Convey("and dphttpclient.Do is called 1 time with the expected method, path and headers", func() {
				checkRequestBase(httpClient, http.MethodGet, "/v1/interactives/123")
			})
		})
	})

	Convey("given a 200 status with empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{
			http.StatusOK,
			[]byte{},
			nil,
		})
		datasetClient := newInteractivesClient(httpClient)

		Convey("when GetInstance is called", func() {
			_, err := datasetClient.GetInteractive(ctx, userAuthToken, serviceAuthToken, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected method, path and headers", func() {
				checkRequestBase(httpClient, http.MethodGet, "/v1/interactives/123")
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := &dphttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("you aint seen me right"))),
				}, nil
			},
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{"/healthcheck"}
			},
		}

		ixClient := newInteractivesClient(httpClient)

		Convey("when GetInstance is called", func() {
			_, err := ixClient.GetInteractive(ctx, userAuthToken, serviceAuthToken, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from interactives api: http://localhost:8080/v1/interactives/123, body: you aint seen me right").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with the expected method, path and headers", func() {
				checkRequestBase(httpClient, http.MethodGet, "/v1/interactives/123")
			})
		})
	})
}

func newInteractivesClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, httpClient)
	interactivesClient := NewWithHealthClient(healthClient, "v1")
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
