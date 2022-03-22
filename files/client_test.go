package files_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {
	timePriorHealthCheck := time.Now()

	Convey("Given the upload service is healthy", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		c := files.NewAPIClient(s.URL)

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

		c := files.NewAPIClient(s.URL)

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
	const filepath = "testing/test.txt"
	const collectionID = "123456789"

	var actualMethod, actualURL, actualContentType string
	var actualContent map[string]string

	Convey("Given a file is uploaded", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actualMethod = r.Method
			actualURL = r.URL.Path
			actualContentType = r.Header.Get("Content-Type")
			json.NewDecoder(r.Body).Decode(&actualContent)

			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then the file collection ID is set", func() {
				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodPatch)
				So(actualContentType, ShouldEqual, "application/json")
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
		c := files.NewAPIClient(s.URL)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldEqual, files.ErrFileNotFound)

			})
		})
	})

	Convey("Given the file already has a collection ID", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldEqual, files.ErrFileAlreadyInCollection)
			})
		})
	})

	Convey("Given files-api has server errors", t, func() {
		errorCode := "CritialError"
		errorDescription := "it is broken"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorBody := fmt.Sprintf(`{"errors": [{"errorCode": "%s", "description": "%s"}]}`, errorCode, errorDescription)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorBody))
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL)

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
		c := files.NewAPIClient(s.URL)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, respContent)
			})
		})
	})

	Convey("given the files api", t, func() {
		c := files.NewAPIClient("broken")

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err, ShouldBeError)
			})
		})
	})
}
