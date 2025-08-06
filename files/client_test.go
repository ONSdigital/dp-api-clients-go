package files_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	. "github.com/smartystreets/goconvey/convey"
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
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When I set the collection ID", func() {
			err := c.SetCollectionID(context.Background(), filepath, collectionID)

			Convey("Then a file not found error should be returned", func() {
				So(err.Error(), ShouldContainSubstring, "unexpected response status code: 418")
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
			w.WriteHeader(http.StatusOK)
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
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))
		defer s.Close()
		c := files.NewAPIClient(s.URL, authHeaderValue)

		Convey("When we publish the collection", func() {

			err := c.PublishCollection(context.Background(), collectionID)

			Convey("Then an error with the response content should be returned", func() {
				So(err.Error(), ShouldContainSubstring, "unexpected response status code: 418")
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

func TestGetFile(t *testing.T) {
	Convey("GetFile called and Files API responds with 200", t, func() {
		Convey("valid file metadata", func() {
			metadata := files.FileMetaData{
				SizeInBytes: uint64(100),
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(metadata)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			filePath := "path/to/file.csv"
			result, err := client.GetFile(context.Background(), filePath, "")

			So(err, ShouldBeNil)
			So(result, ShouldResemble, metadata)
		})

		Convey("valid file metadata through v1 endpoint (dp-api-router)", func() {
			metadata := files.FileMetaData{
				SizeInBytes: uint64(100),
			}

			server := mockServerWithVersionedEndpoint(metadata, "/v1")

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			client.Version = "v1"

			filePath := "path/to/file.csv"
			result, err := client.GetFile(context.Background(), filePath, "")

			So(err, ShouldBeNil)
			So(result, ShouldResemble, metadata)
		})

		Convey("valid file metadata through v2 endpoint returns an error", func() {
			metadata := files.FileMetaData{
				SizeInBytes: uint64(100),
			}

			server := mockServerWithVersionedEndpoint(metadata, "/v1")

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			client.Version = "v2"

			filePath := "path/to/file.csv"
			_, err := client.GetFile(context.Background(), filePath, "")

			So(err, ShouldBeError)
		})

		Convey("invalid JSON", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "<invalid JSON>")
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			filePath := "path/to/file.csv"
			_, err := client.GetFile(context.Background(), filePath, "auth-token")

			So(err, ShouldBeError)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})
	})

	Convey("GetFile call errors", t, func() {
		Convey("known errors that return JSON responses", func() {
			Convey("404 file does not exist", func() {
				expectedCode := "FileNotRegistered"
				expectedDescription := "file not registered"
				server := newMockFilesAPIServerWithError(http.StatusNotFound, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				_, err := client.GetFile(context.Background(), "path/to/file.csv", "auth-token")

				So(err, ShouldBeError)
				So(err.Error(), ShouldEqual, fmt.Sprintf("%s: %s", expectedCode, expectedDescription))
			})

			Convey("500 internal server error", func() {
				expectedCode := "InternalError"
				expectedDescription := "no space on disk"
				server := newMockFilesAPIServerWithError(http.StatusInternalServerError, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				_, err := client.GetFile(context.Background(), "path/to/file.csv", "auth-token")

				So(err, ShouldBeError)
				So(err.Error(), ShouldEqual, fmt.Sprintf("%s: %s: %s", files.ErrServer, expectedCode, expectedDescription))
			})

			Convey("invalid JSON error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, "<invalid JSON>")
				}))

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)

				filePath := "path/to/file.csv"
				_, err := client.GetFile(context.Background(), filePath, "auth-token")

				So(err, ShouldBeError)
				So(err.Error(), ShouldContainSubstring, "invalid character")
			})
		})

		Convey("unknown error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			filePath := "path/to/file.csv"
			_, err := client.GetFile(context.Background(), filePath, "auth-token")

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "unexpected response status code: 418")
		})

		Convey("HTTP client error", func() {
			hCli := health.Client{URL: "broken", Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			_, err := client.GetFile(context.Background(), "path/to/file.txt", "auth-token")
			So(err, ShouldBeError)
		})
	})

	Convey("GetFile authorises requests to Files API", t, func() {
		Convey("adds a service token to the header", func() {
			expectedToken := "auth-token"
			expectedBearerToken := fmt.Sprintf("Bearer %s", expectedToken)

			var token string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				token = req.Header.Get("Authorization")
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			client.GetFile(context.Background(), "path/to/file.csv", expectedToken)
			So(token, ShouldEqual, expectedBearerToken)
		})

		Convey("returns an error if unauthorised", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			_, err := client.GetFile(context.Background(), "path/to/file.csv", "invalid-token")
			So(err, ShouldEqual, files.ErrNotAuthorized)
		})
	})
}

func TestRegisterFile(t *testing.T) {
	Convey("RegisterFile happy path", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
		client := files.NewWithHealthClient(&hCli)

		err := client.RegisterFile(context.Background(), files.FileMetaData{})

		So(err, ShouldBeNil)
	})

	Convey("RegisterFile call errors", t, func() {
		Convey("internal server error", func() {
			Convey("with valid JSON description", func() {
				expectedCode := "InternalError"
				expectedDescription := "no space on disk"
				server := newMockFilesAPIServerWithError(http.StatusInternalServerError, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(err.Error(), ShouldEqual, fmt.Sprintf("%s: %s: %s", files.ErrServer, expectedCode, expectedDescription))
			})

			Convey("with invalid JSON description", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, "<invalid JSON>")
				}))

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)

				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(err.Error(), ShouldContainSubstring, "invalid character")
			})
		})

		Convey("bad request", func() {
			Convey("duplicate file", func() {
				expectedCode := "DuplicateFileError"
				expectedDescription := ""
				server := newMockFilesAPIServerWithError(http.StatusBadRequest, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(err, ShouldBeError, files.ErrFileAlreadyRegistered)
				So(errors.Is(err, files.ErrBadRequest), ShouldBeTrue)
				So(err.Error(), ShouldEqual, "bad request: file already registered")
			})

			Convey("validation error", func() {
				expectedCode := "ValidationError"
				expectedDescription := "path not provided"
				server := newMockFilesAPIServerWithError(http.StatusBadRequest, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(errors.Is(err, files.ErrBadRequest), ShouldBeTrue)
				So(errors.Is(err, files.ErrValidationError), ShouldBeTrue)
				So(err.Error(), ShouldEqual, "bad request: validation error: path not provided")
			})

			Convey("unknown code", func() {
				expectedCode := "BizarreError"
				expectedDescription := "path not provided"
				server := newMockFilesAPIServerWithError(http.StatusBadRequest, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(errors.Is(err, files.ErrBadRequest), ShouldBeTrue)
				So(errors.Is(err, files.ErrUnknown), ShouldBeTrue)
				So(err.Error(), ShouldEqual, "bad request: unknown error: BizarreError: path not provided")
			})

			Convey("invalid JSON", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(w, "<invalid JSON>")
				}))

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.RegisterFile(context.Background(), files.FileMetaData{})

				So(err, ShouldBeError)
				So(errors.Is(err, files.ErrBadRequest), ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, "invalid character")
			})
		})

		Convey("unknown error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			err := client.RegisterFile(context.Background(), files.FileMetaData{})

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "unexpected response status code: 418")
		})

		Convey("HTTP client error", func() {
			hCli := health.Client{URL: "broken", Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			err := client.RegisterFile(context.Background(), files.FileMetaData{})

			So(err, ShouldBeError)
		})
	})

	Convey("RegisterFile authorises requests to Files API", t, func() {
		Convey("adds a service token to the header", func() {
			expectedToken := "auth-token"
			expectedBearerToken := fmt.Sprintf("Bearer %s", expectedToken)

			var token string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				token = req.Header.Get("Authorization")
				w.WriteHeader(http.StatusCreated)
			}))

			client := files.NewAPIClient(server.URL, expectedToken)

			err := client.RegisterFile(context.Background(), files.FileMetaData{})

			So(err, ShouldBeNil)
			So(token, ShouldEqual, expectedBearerToken)
		})

		Convey("returns an error if unauthorised", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			err := client.RegisterFile(context.Background(), files.FileMetaData{})
			So(err, ShouldEqual, files.ErrNotAuthorized)
		})

		Convey("resource conflict", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusConflict)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			err := client.RegisterFile(context.Background(), files.FileMetaData{})

			So(err, ShouldEqual, files.ErrConflict)
		})
	})
}

func TestPatchFile(t *testing.T) {
	Convey("PatchFile happy path", t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
		client := files.NewWithHealthClient(&hCli)

		err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

		So(err, ShouldBeNil)
	})

	Convey("PatchFile call errors", t, func() {
		Convey("internal server error", func() {
			Convey("with valid JSON description", func() {
				expectedCode := "InternalError"
				expectedDescription := "no space on disk"
				server := newMockFilesAPIServerWithError(http.StatusInternalServerError, expectedCode, expectedDescription)

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)
				err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

				So(err, ShouldBeError)
				So(err.Error(), ShouldEqual, fmt.Sprintf("%s: %s: %s", files.ErrServer, expectedCode, expectedDescription))
			})

			Convey("with invalid JSON description", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, "<invalid JSON>")
				}))

				hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
				client := files.NewWithHealthClient(&hCli)

				err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

				So(err, ShouldBeError)
				So(err.Error(), ShouldContainSubstring, "invalid character")
			})
		})

		Convey("bad request", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeError)
			So(err, ShouldBeError, files.ErrFileAlreadyInCollection)
			So(err.Error(), ShouldEqual, "file collection ID already set")

		})

		Convey("file not publishable", func() {
			expectedCode := "FileNotPublishable"
			server := newMockFilesAPIServerWithError(http.StatusConflict, expectedCode, "")

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "bad request: file already registered")
		})

		Convey("file is in invalid state", func() {
			expectedCode := "unspecified invalid state"
			server := newMockFilesAPIServerWithError(http.StatusConflict, expectedCode, "")

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeError, files.ErrInvalidState)
		})

		Convey("unknown error", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeError)
			So(err.Error(), ShouldEqual, "unexpected response status code: 418")
		})

		Convey("HTTP client error", func() {
			hCli := health.Client{URL: "broken", Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)
			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeError)
		})
	})

	Convey("PatchFile authorises requests to Files API", t, func() {
		Convey("adds a service token to the header", func() {
			expectedToken := "auth-token"
			expectedBearerToken := fmt.Sprintf("Bearer %s", expectedToken)

			var token string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				token = req.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
			}))

			client := files.NewAPIClient(server.URL, expectedToken)

			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})

			So(err, ShouldBeNil)
			So(token, ShouldEqual, expectedBearerToken)
		})

		Convey("returns an error if unauthorised", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}))

			hCli := health.Client{URL: server.URL, Client: &dphttp.Client{}}
			client := files.NewWithHealthClient(&hCli)

			err := client.PatchFile(context.Background(), "a.txt", files.FilePatch{})
			So(err, ShouldEqual, files.ErrNotAuthorized)
		})
	})
}

func mockServerWithVersionedEndpoint(metadata files.FileMetaData, version string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqURL, _ := url.Parse(req.RequestURI)
		if reqURL.Path[0:3] == version {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(metadata)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newMockFilesAPIServerWithError(expectedStatus int, expectedCode, expectedError string) *httptest.Server {
	jsonError := dperrors.JsonErrors{
		Errors: []dperrors.JsonError{
			{Code: expectedCode, Description: expectedError},
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(expectedStatus)
		json.NewEncoder(w).Encode(jsonError)
	}))
}
