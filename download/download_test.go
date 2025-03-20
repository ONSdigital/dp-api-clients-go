package download_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/download"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/v3/request"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	rootpath        = "downloads-new"
	filepath        = "testing/test.txt"
	authHeaderValue = "a-service-client-auth-token"
	actualContent   = "some-content"
)

var (
	actualMethod, actualURL, actualAuthHeaderValue string
)

func TestHealthCheck(t *testing.T) {
	timePriorHealthCheck := time.Now()

	Convey("Given the download service is healthy", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		c := download.NewAPIClient(s.URL, "")

		Convey("When we check that state of the service", func() {
			state := health.CreateCheckState("testing")
			c.Checker(context.Background(), &state)

			Convey("Then the health check should be successful", func() {
				So(state.Status(), ShouldEqual, healthcheck.StatusOK)
				So(state.StatusCode(), ShouldEqual, 200)
				So(state.Message(), ShouldContainSubstring, "is ok")
			})

			Convey("And the timestamps are logged appropriately", func() {
				So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*state.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(state.LastFailure(), ShouldBeNil)
			})
		})
	})

	Convey("Given the download service is failing", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
		defer s.Close()

		c := download.NewAPIClient(s.URL, "")

		Convey("When we check the state of the service", func() {
			state := health.CreateCheckState("testing")
			c.Checker(context.Background(), &state)

			Convey("Then the health check should be successful", func() {
				So(state.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(state.StatusCode(), ShouldEqual, 500)
				So(state.Message(), ShouldContainSubstring, "unavailable or non-functioning")
			})

			Convey("And the timestamps are logged appropriately", func() {
				So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(state.LastSuccess(), ShouldBeNil)
				So(*state.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})
		})
	})
}

func TestDownload(t *testing.T) {

	Convey("Given a file is available for download", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
			actualURL = r.URL.Path
			actualAuthHeaderValue = r.Header.Get(dprequest.AuthHeaderKey)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(actualContent))
		}))
		defer s.Close()

		c := download.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then the response is as expected", func() {
				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodGet)
				So(actualAuthHeaderValue, ShouldEqual, fmt.Sprintf("Bearer %s", authHeaderValue))
				So(actualURL, ShouldEqual, fmt.Sprintf("/%s/%s", rootpath, filepath))
				So(actualContent, ShouldEqual, content)
			})
		})
	})

	Convey("Given a file responds as expected from a redirect", t, func() {
		redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
			actualURL = r.URL.Path
			actualAuthHeaderValue = r.Header.Get(dprequest.AuthHeaderKey)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(actualContent))
		}))
		defer redirect.Close()

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, redirect.URL, http.StatusMovedPermanently)
		}))
		defer s.Close()

		c := download.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then the response is as expected", func() {
				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodGet)
				So(actualURL, ShouldEqual, "/")
				So(actualContent, ShouldEqual, content)
				So(actualAuthHeaderValue, ShouldEqual, "Bearer a-service-client-auth-token")
			})
		})
	})

	Convey("Given there no file available", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer s.Close()

		c := download.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then a well formatted error should be returned", func() {
				So(err, ShouldHaveSameTypeAs, &dperrors.Error{})
			})

			Convey("And the expected error is returned", func() {
				dperr := err.(*dperrors.Error)
				So(dperr.Error(), ShouldBeEmpty)
				So(dperr.Code(), ShouldEqual, http.StatusNotFound)
			})

			Convey("And the content should be empty", func() {
				So(content, ShouldBeEmpty)
			})
		})
	})

	Convey("Given the service auth token is not valid", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer s.Close()

		c := download.NewAPIClient(s.URL, "not-valid-token")

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then a well formatted error should be returned", func() {
				So(err, ShouldHaveSameTypeAs, &dperrors.Error{})
			})

			Convey("And the expected error is returned", func() {
				dperr := err.(*dperrors.Error)
				So(dperr.Error(), ShouldBeEmpty)
				So(dperr.Code(), ShouldEqual, http.StatusForbidden)
			})

			Convey("And the content should be empty", func() {
				So(content, ShouldBeEmpty)
			})
		})
	})

	Convey("Given download service has downstream errors", t, func() {
		errorCode := "CriticalError"
		errorDescription := "it is broken"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorBody := fmt.Sprintf(`{"errors": [{"errorCode": "%s", "description": "%s"}]}`, errorCode, errorDescription)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorBody))
		}))
		defer s.Close()

		c := download.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then a well formatted error should be returned", func() {
				So(err, ShouldHaveSameTypeAs, &dperrors.Error{})
			})

			Convey("And the expected error is returned", func() {
				dperr := err.(*dperrors.Error)
				So(dperr.Error(), ShouldContainSubstring, errorCode)
				So(dperr.Error(), ShouldContainSubstring, errorDescription)
				So(dperr.Code(), ShouldEqual, http.StatusInternalServerError)
			})

			Convey("And the content should be empty", func() {
				So(content, ShouldBeEmpty)
			})
		})
	})

	Convey("Given download service is configured incorrectly", t, func() {
		c := download.NewAPIClient("broken", authHeaderValue)

		Convey("When I download", func() {
			resp, err := c.Download(context.Background(), filepath)
			content := readAndClose(resp)

			Convey("Then a well formatted error should be returned", func() {
				So(err, ShouldHaveSameTypeAs, &dperrors.Error{})
			})

			Convey("And the expected error is returned", func() {
				dperr := err.(*dperrors.Error)
				So(dperr.Error(), ShouldContainSubstring, "broken")
				So(dperr.Code(), ShouldEqual, http.StatusInternalServerError)
			})

			Convey("And the content should be empty", func() {
				So(content, ShouldBeEmpty)
			})
		})
	})
}

func readAndClose(response *download.Response) string {
	if response == nil {
		return ""
	}
	content, _ := io.ReadAll(response.Content)
	response.Content.Close()
	return string(content)
}
