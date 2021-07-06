package identity

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/headers"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	url                = "/whatever"
	florenceToken      = "roundabout"
	callerAuthToken    = "YourClaimToBeWhoYouAre"
	callerIdentifier   = "externalCaller"
	userIdentifier     = "fred@ons.gov.uk"
	zebedeeURL         = "http://localhost:8082"
	expectedZebedeeURL = zebedeeURL + "/identity"
)

func TestHandler_NoAuth(t *testing.T) {

	Convey("Given a request with no auth headers", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		httpClient := newMockHTTPClient()
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, "", "")

			Convey("Then the downstream HTTP handler should not be called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
				So(err, ShouldBeNil)
				So(authFailure.Error(), ShouldContainSubstring, "no headers set on request: "+errUnableToIdentifyRequest.Error())
				So(ctx, ShouldNotBeNil)
				So(dprequest.IsUserPresent(ctx), ShouldBeFalse)
				So(dprequest.IsCallerPresent(ctx), ShouldBeFalse)
			})

			Convey("Then the returned code should be 401", func() {
				So(status, ShouldEqual, http.StatusUnauthorized)
			})
		})
	})
}

func TestHandler_IdentityServiceError(t *testing.T) {

	Convey("Given a request with a florence token, and a mock client that returns an error", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		expectedError := errors.New("broken")
		httpClient := getClientReturningError(expectedError)
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service was called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
			})

			Convey("Then the error and no context is returned", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldEqual, expectedError)
				So(status, ShouldNotEqual, http.StatusOK)
				So(ctx, ShouldNotBeNil)
				So(dprequest.IsUserPresent(ctx), ShouldBeFalse)
				So(dprequest.IsCallerPresent(ctx), ShouldBeFalse)
			})
		})
	})
}

func TestHandler_IdentityServiceErrorResponseCode(t *testing.T) {
	Convey("Given a request with a florence token, and mock client that returns a non-200 response", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		body := httpmocks.NewReadCloserMock([]byte{}, nil)
		authResp := httpmocks.NewResponseMock(body, http.StatusNotFound)
		httpClient := newMockHTTPClient()
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return authResp, nil
		}
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedZebedeeURL)
			})

			Convey("Then there is no error but the response code matches the identity service", func() {
				So(authFailure.Error(), ShouldContainSubstring, "unexpected status code returned from AuthAPI: "+errUnableToIdentifyRequest.Error())
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusNotFound)
				So(ctx, ShouldNotBeNil)
				So(dprequest.IsUserPresent(ctx), ShouldBeFalse)
				So(dprequest.IsCallerPresent(ctx), ShouldBeFalse)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_florenceToken(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		httpClient, body, _ := getClientReturningIdentifier(t, userIdentifier)
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusOK)
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler returned no error and expected context", func() {
				So(ctx, ShouldNotBeNil)
				So(dprequest.Caller(ctx), ShouldEqual, userIdentifier)
				So(dprequest.User(ctx), ShouldEqual, userIdentifier)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a request with a florence token as a cookie, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		httpClient, body, _ := getClientReturningIdentifier(t, userIdentifier)
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusOK)
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the downstream HTTP handler returned no error and expected context", func() {
				So(ctx, ShouldNotBeNil)
				So(dprequest.Caller(ctx), ShouldEqual, userIdentifier)
				So(dprequest.User(ctx), ShouldEqual, userIdentifier)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_InvalidIdentityResponse(t *testing.T) {

	Convey("Given a request with a florence token, and mock client that returns invalid response JSON", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		b := []byte("{ invalid JSON")
		body := httpmocks.NewReadCloserMock(b, nil)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient()
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, nil
		}
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the response is set as expected", func() {
				So(authFailure, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid character 'i' looking for beginning of object key string")
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(ctx, ShouldNotBeNil)
				So(dprequest.Caller(ctx), ShouldBeEmpty)
				So(dprequest.User(ctx), ShouldBeEmpty)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_ReadBodyError(t *testing.T) {
	Convey("Given a ioutil.ReadAll returns an error when reading the response body", t, func() {
		req := httptest.NewRequest("GET", url, nil)

		expectedErr := errors.New("cause i'm tnt i'm dynamite tnt and i'll win the fight")
		body := httpmocks.NewReadCloserMock(nil, expectedErr)
		resp := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient()
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return resp, nil
		}
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			_, status, _, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is called as expected", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("And the expected status code and error is returned", func() {
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(err, ShouldResemble, expectedErr)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_NewAuthRequestError(t *testing.T) {
	Convey("Given creating a new auth request returns an error", t, func() {
		req := httptest.NewRequest("GET", url, nil)
		httpClient := newMockHTTPClient()
		httpClient.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return nil, nil
		}
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", "£$%^&*(((((", httpClient))

		Convey("When CheckRequest is called", func() {

			_, status, _, err := idClient.CheckRequest(req, florenceToken, "")

			Convey("Then the identity service is not called", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 0)
			})

			Convey("And the expected status code and error is returned", func() {
				So(status, ShouldEqual, http.StatusInternalServerError)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestHandler_authToken(t *testing.T) {

	Convey("Given a request with an auth token, and mock client that returns 200", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			dprequest.UserHeaderKey: {userIdentifier},
		}
		httpClient, body, _ := getClientReturningIdentifier(t, callerIdentifier)
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, "", callerAuthToken)
			So(err, ShouldBeNil)
			So(authFailure, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)

			Convey("Then the identity service is called as expected", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.UserHeaderKey], ShouldHaveLength, 0)
				So(zebedeeReq.Header[dprequest.AuthHeaderKey], ShouldHaveLength, 1)
				actual, err := headers.GetServiceAuthToken(zebedeeReq)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, callerAuthToken)
			})

			Convey("Then the downstream HTTP handler request has the expected context values", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)

				So(ctx, ShouldNotBeNil)
				So(dprequest.IsCallerPresent(ctx), ShouldBeTrue)
				So(dprequest.IsUserPresent(ctx), ShouldBeTrue)
				So(dprequest.Caller(ctx), ShouldEqual, callerIdentifier)
				So(dprequest.User(ctx), ShouldEqual, userIdentifier)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestHandler_bothTokens(t *testing.T) {

	Convey("Given a request with both a florence token and service token", t, func() {

		req := httptest.NewRequest("GET", url, nil)
		req.Header = map[string][]string{
			dprequest.FlorenceHeaderKey: {florenceToken},
			dprequest.AuthHeaderKey:     {callerAuthToken},
		}
		httpClient, body, _ := getClientReturningIdentifier(t, userIdentifier)
		idClient := NewWithHealthClient(healthcheck.NewClientWithClienter("", zebedeeURL, httpClient))

		Convey("When CheckRequest is called", func() {

			ctx, status, authFailure, err := idClient.CheckRequest(req, florenceToken, callerAuthToken)
			So(err, ShouldBeNil)
			So(authFailure, ShouldBeNil)
			So(status, ShouldEqual, http.StatusOK)

			Convey("Then the identity service is called as expected - verifying florence", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 1)
				zebedeeReq := httpClient.DoCalls()[0].Req
				So(zebedeeReq.URL.String(), ShouldEqual, expectedZebedeeURL)
				So(zebedeeReq.Header[dprequest.FlorenceHeaderKey][0], ShouldEqual, florenceToken)
			})

			Convey("Then the context returns with expected values", func() {
				So(ctx, ShouldNotBeNil)
				So(dprequest.IsUserPresent(ctx), ShouldBeTrue)
				So(dprequest.User(ctx), ShouldEqual, userIdentifier)
				So(dprequest.Caller(ctx), ShouldEqual, userIdentifier)
			})

			Convey("And Auth API response body is closed", func() {
				So(body.IsClosed, ShouldBeTrue)
			})
		})
	})
}

func TestSplitTokens(t *testing.T) {
	Convey("Given a service token and an empty florence token", t, func() {
		florenceToken := ""
		serviceToken := "Bearer 123456789"

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 2, hasPrefix: true, tokenPart: "456789"})
				So(logData["florence_token"], ShouldBeNil)
			})
		})
	})

	Convey("Given a florence token and an empty service token", t, func() {
		florenceToken := "987654321"
		serviceToken := ""

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "654321"})
				So(logData["auth_token"], ShouldBeNil)
			})
		})
	})

	Convey("Given a florence token and service token", t, func() {
		florenceToken := "987654321"
		serviceToken := "Bearer 123456789"

		Convey("When we pass both tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "654321"})
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 2, hasPrefix: true, tokenPart: "456789"})
			})
		})
	})

	Convey("Given a small service token", t, func() {
		florenceToken := "54321"
		serviceToken := "Bearer A 12"

		Convey("When we pass the tokens into splitTokens function", func() {
			logData := log.Data{}
			splitTokens(florenceToken, serviceToken, logData)

			Convey("Then the token objects are returned with the expected values", func() {
				So(logData["florence_token"], ShouldResemble, tokenObject{numberOfParts: 1, hasPrefix: false, tokenPart: "321"})
				So(logData["auth_token"], ShouldResemble, tokenObject{numberOfParts: 3, hasPrefix: true, tokenPart: "2"})
			})
		})
	})

}

func newMockHTTPClient() *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func getClientReturningIdentifier(t *testing.T, id string) (*dphttp.ClienterMock, *httpmocks.ReadCloserMock, *http.Response) {
	b := httpmocks.GetEntityBytes(t, &dprequest.IdentityResponse{Identifier: id})
	body := httpmocks.NewReadCloserMock(b, nil)
	resp := httpmocks.NewResponseMock(body, http.StatusOK)
	cli := newMockHTTPClient()
	cli.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return resp, nil
	}
	return cli, body, resp
}

func getClientReturningError(err error) *dphttp.ClienterMock {
	cli := newMockHTTPClient()
	cli.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return nil, err
	}
	return cli
}
