package filterflex_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/filterflex"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	. "github.com/smartystreets/goconvey/convey"
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
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
func TestForwardRequest(t *testing.T) {
	Convey("Given a client intialised to a mock filterFlexAPI that returns details of the incoming request", t, func() {

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(b) == 0 {
				b = []byte("{nil}")
			}

			fmt.Fprintf(w, "Request for: %s %s Body: %s Header: %s", r.Method, r.URL.String(), string(b), r.Header.Get("Expected"))
		}))

		defer svr.Close()

		client := filterflex.New(filterflex.Config{
			HostURL: svr.URL,
		})

		Convey("when ForwardRequest is called with a GET request with a nil body", func() {
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Expected", "Value")

			resp, err := client.ForwardRequest(req)
			So(err, ShouldBeNil)

			b, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			expected := "Request for: GET /foo Body: {nil} Header: Value"
			So(string(b), ShouldResemble, expected)
		})

		Convey("when ForwardRequest is called with a GET request with a query param", func() {
			req := httptest.NewRequest(http.MethodGet, "/foo?limit=1", nil)
			req.Header.Set("Expected", "Value")

			resp, err := client.ForwardRequest(req)
			So(err, ShouldBeNil)

			b, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			expected := "Request for: GET /foo?limit=1 Body: {nil} Header: Value"
			So(string(b), ShouldResemble, expected)
		})

		Convey("when ForwardRequest is called with a POST request with a body", func() {
			req := httptest.NewRequest(http.MethodPost, "/bar", bytes.NewReader([]byte("I am body")))
			req.Header.Set("Expected", "OtherValue")

			resp, err := client.ForwardRequest(req)
			So(err, ShouldBeNil)

			b, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			expected := "Request for: POST /bar Body: I am body Header: OtherValue"
			So(string(b), ShouldResemble, expected)
		})
	})
}

func TestDeleteOptionsOption(t *testing.T) {
	const userAuthToken = "userAuth"
	const serviceAuthToken = "serviceAuth"
	const sentETag = "sentETag"
	const receivedETag = "testETag"

	Convey("Given a valid request", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{
			http.StatusNoContent,
			ioutil.NopCloser(bytes.NewReader(nil)),
			map[string]string{"ETag": receivedETag},
		})

		cfg := filterflex.Config{
			HostURL: "http://test.test:2000",
		}
		cliH := health.NewClientWithClienter("", "http://test.test:2000", httpClient)
		client := filterflex.NewWithHealthClient(cfg, cliH)
		delOption := filterflex.GetDeleteOptionInput{
			FilterID:  "filter_id",
			Dimension: "dimension",
			Option:    "option",
			IfMatch:   sentETag,
			AuthHeaders: filterflex.AuthHeaders{
				UserAuthToken:    userAuthToken,
				ServiceAuthToken: serviceAuthToken,
			},
		}
		eTag, err := client.DeleteOption(context.Background(), delOption)
		So(err, ShouldBeNil)
		So(eTag, ShouldEqual, receivedETag)

		Convey("it should call the delete endpoint, serializing the delete query", func() {
			calls := httpClient.DoCalls()
			So(calls, ShouldNotBeEmpty)
			So(calls[0].Req.URL.String(), ShouldEqual, "http://test.test:2000/filters/filter_id/dimensions/dimension/options/option")
			So(calls[0].Req, shouldHaveAuthHeaders, userAuthToken, serviceAuthToken, sentETag)
		})
	})
}

// shouldHaveAuthHeaders is a GoConvey matcher that asserts the values of the
// auth headers on a request match the expected values.
// Usage: `So(request, shouldHaveAuthHeaders, "userToken", "serviceToken")`
func shouldHaveAuthHeaders(actual interface{}, expected ...interface{}) string {
	req, ok := actual.(*http.Request)
	if !ok {
		return "expected to find http.Request"
	}

	if len(expected) != 3 {
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
	expETagHeader, ok := expected[2].(string)
	if !ok {
		return "ETag must be a string"
	}

	florenceToken := req.Header.Get("X-Florence-Token")
	if florenceToken != expUserHeader {
		return fmt.Sprintf("expected X-Florence-Token value %s, got %s", florenceToken, expUserHeader)
	}

	svcHeader := req.Header.Get("Authorization")
	if svcHeader != fmt.Sprintf("Bearer %s", expSvcHeader) {
		return fmt.Sprintf("expected Authorization value %s, got %s", svcHeader, expSvcHeader)
	}

	eTag := req.Header.Get("If-Match")
	if eTag != expETagHeader {
		return fmt.Sprintf("expected Etag value %s, got %s", eTag, expETagHeader)
	}
	return ""
}
