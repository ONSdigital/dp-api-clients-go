package files_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dprequest "github.com/ONSdigital/dp-net/request"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	filepath        = "testing/test.txt"
	collectionID    = "123456789"
	authHeaderValue = "a-service-client-auth-token"
)

var actualMethod, actualURL, actualContentType, actualAuthHeaderValue string
var actualContent map[string]string

func TestHealthCheck(t *testing.T) {
	timePriorHealthCheck := time.Now()

	Convey("Given the upload service is healthy", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		c := files.NewAPIClient(s.URL, "does-not-matter")

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

	Convey("Given the upload service is failing", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
		defer s.Close()

		c := files.NewAPIClient(s.URL, "does-not-matter")

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

func TestSetCollectionID(t *testing.T) {

	Convey("Given a file is uploaded", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
			actualURL = r.URL.Path
			actualContentType = r.Header.Get("Content-Type")
			actualAuthHeaderValue = r.Header.Get(dprequest.AuthHeaderKey)
			json.NewDecoder(r.Body).Decode(&actualContent)

			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then the file collection ID is set", func() {
				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodPatch)
				So(actualContentType, ShouldEqual, "application/json")
				So(actualAuthHeaderValue, ShouldEqual, fmt.Sprintf("Bearer %s", authHeaderValue))
				So(actualURL, ShouldEqual, fmt.Sprintf("/files/%s", filepath))
				So(actualContent["collection_id"], ShouldEqual, collectionID)

			})
		})
	})

	Convey("Given there no file uploaded", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldEqual, files.ErrFileNotFound)

			})
		})
	})

	Convey("Given the service auth token is not valid", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, "not-valid-token")

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a not authorised error should be returned", func() {
				So(err, ShouldEqual, files.ErrNotAuthorized)

			})
		})
	})

	Convey("Given the file already has a collection ID", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldEqual, files.ErrFileAlreadyInCollection)
			})
		})
	})

	Convey("Given files-api has server errors", t, func() {
		errorCode := "CriticalError"
		errorDescription := "it is broken"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorBody := fmt.Sprintf(`{"errors": [{"errorCode": "%s", "description": "%s"}]}`, errorCode, errorDescription)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorBody))
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("%s: %s", errorCode, errorDescription))
			})
		})
	})

	Convey("Given the file already has a collection ID", t, func() {
		respContent := "i'm a little tea pot..."
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(respContent))
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, respContent)
			})
		})
	})

	Convey("given the files api", t, func() {
		c := files.NewAPIClient("broken", authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestPublishCollection(t *testing.T) {
	Convey("There are file in the collection to be published", t, func() {

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
			actualURL = r.URL.Path
			actualAuthHeaderValue = r.Header.Get(dprequest.AuthHeaderKey)
			w.WriteHeader(http.StatusCreated)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("The collection is published", func() {
				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodPatch)
				So(actualURL, ShouldEqual, fmt.Sprintf("/collection/%s", collectionID))
				So(actualAuthHeaderValue, ShouldEqual, fmt.Sprintf("Bearer %s", authHeaderValue))
			})
		})
	})

	Convey("The files are not in an UPLOADED state", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("The invalid state error is returned", func() {
				So(err, ShouldEqual, files.ErrInvalidState)
			})
		})
	})
	Convey("There are no files in the collection", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("The a no files in collection error is returned", func() {
				So(err, ShouldEqual, files.ErrNoFilesInCollection)
			})
		})
	})
	Convey("The auth token is not valid", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, "not-valid-auth")

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("an not authorized error should be returned", func() {
				So(err, ShouldEqual, files.ErrNotAuthorized)
			})
		})
	})

	Convey("There is a server error", t, func() {
		errorCode := "CriticalError"
		errorDescription := "it is broken"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorBody := fmt.Sprintf(`{"errors": [{"errorCode": "%s", "description": "%s"}]}`, errorCode, errorDescription)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorBody))
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("Then an error container the JSON Error content should be returned", func() {
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("%s: %s", errorCode, errorDescription))
			})
		})
	})

	Convey("There is an expected response", t, func() {
		respContent := "Testing Testing 123"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(respContent))
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("Then an error with the response content should be returned", func() {
				So(err.Error(), ShouldContainSubstring, fmt.Sprintf("unexpected error: %s", respContent))
			})
		})
	})

	Convey("There is an error connecting to files-api", t, func() {
		c := files.NewAPIClient("broken", authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("An error should be returned", func() {
				So(err, ShouldBeError)
			})
		})
	})
}