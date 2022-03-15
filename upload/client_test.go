package upload_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	actualContent              string
	actualHasCollectionID      bool
	actualCollectionId         string
	actualResumableFilename    string
	actualPath                 string
	actualIsPublishable        string
	actualTitle                string
	actualResumableTotalSize   string
	actualResumableType        string
	actualLicence              string
	actualLicenceURL           string
	actualResumableChunkNumber string
	actualResumableTotalChunks string

	actualMethod     string
	numberOfAPICalls int

	collectionID = "123456"
)

const (
	filename      = "file.txt"
	path          = "data/file.txt"
	isPublishable = false
	title         = "Information about shoe size"
	fileType      = "text/plain"
	license       = "MIT"
	licenseURL    = "https://opensource.org/licenses/MIT"
)

func TestHealthCheck(t *testing.T) {
	timePriorHealthCheck := time.Now()

	Convey("Given the upload service is healthy", t, func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		c := upload.NewAPIClient(s.URL)

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

		c := upload.NewAPIClient(s.URL)

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

func TestUpload(t *testing.T) {
	Convey("Given the upload service is running", t, func() {
		actualContent = ""
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			extractFields(r)

			w.WriteHeader(http.StatusCreated)
		}))
		defer s.Close()
		c := upload.NewAPIClient(s.URL)

		Convey("And the file is a single chunk", func() {
			fileContent := "testing"
			f := io.NopCloser(strings.NewReader(fileContent))

			Convey("When I upload the single-chunk file with metadata containing a collection ID", func() {
				metadata := upload.Metadata{
					CollectionID:  &collectionID,
					FileName:      filename,
					Path:          path,
					IsPublishable: isPublishable,
					Title:         title,
					FileSizeBytes: int64(len(fileContent)),
					FileType:      fileType,
					License:       license,
					LicenseURL:    licenseURL,
				}

				numberOfAPICalls = 0
				err := c.Upload(context.Background(), f, metadata)

				Convey("Then the file is successfully uploaded", func() {
					So(err, ShouldBeNil)
					So(actualMethod, ShouldEqual, http.MethodPost)
					So(actualContent, ShouldEqual, fileContent)
				})

				Convey("And the resumable data was calculated", func() {
					So(actualResumableFilename, ShouldEqual, filename)
					So(actualResumableTotalSize, ShouldEqual, fmt.Sprintf("%d", len(fileContent)))
					So(actualResumableChunkNumber, ShouldEqual, "1")
					So(actualResumableTotalChunks, ShouldEqual, "1")
					So(actualResumableType, ShouldEqual, fileType)
				})

				Convey("And the file metadata is sent with the file", func() {
					So(actualCollectionId, ShouldEqual, collectionID)
					So(actualPath, ShouldEqual, path)
					So(actualIsPublishable, ShouldEqual, strconv.FormatBool(isPublishable))
					So(actualTitle, ShouldEqual, title)
					So(actualLicence, ShouldEqual, license)
					So(actualLicenceURL, ShouldEqual, licenseURL)
				})
			})

			Convey("When I upload the single-chunk file with metadata not containing a collection ID", func() {
				metadata := upload.Metadata{
					FileName:      filename,
					Path:          path,
					IsPublishable: isPublishable,
					Title:         title,
					FileSizeBytes: int64(len(fileContent)),
					FileType:      fileType,
					License:       license,
					LicenseURL:    licenseURL,
				}

				numberOfAPICalls = 0
				err := c.Upload(context.Background(), f, metadata)

				Convey("Then the file is successfully uploaded", func() {
					So(err, ShouldBeNil)
					So(actualMethod, ShouldEqual, http.MethodPost)
					So(actualContent, ShouldEqual, fileContent)
				})

				Convey("And the resumable data was calculated", func() {
					So(actualResumableFilename, ShouldEqual, filename)
					So(actualResumableTotalSize, ShouldEqual, fmt.Sprintf("%d", len(fileContent)))
					So(actualResumableChunkNumber, ShouldEqual, "1")
					So(actualResumableTotalChunks, ShouldEqual, "1")
					So(actualResumableType, ShouldEqual, fileType)
				})

				Convey("And the file metadata is sent with the file", func() {
					So(actualHasCollectionID, ShouldBeFalse)
					So(actualPath, ShouldEqual, path)
					So(actualIsPublishable, ShouldEqual, strconv.FormatBool(isPublishable))
					So(actualTitle, ShouldEqual, title)
					So(actualLicence, ShouldEqual, license)
					So(actualLicenceURL, ShouldEqual, licenseURL)
				})
			})
		})

		Convey("And the file is multiple chunks", func() {
			expectedContentLength := 6 * 1024 * 1024

			contentBytes := make([]byte, expectedContentLength)
			rand.Read(contentBytes)

			fileContent := hex.EncodeToString(contentBytes)
			f := io.NopCloser(strings.NewReader(fileContent))

			Convey("When I upload the multi-chunk file with metadata containing a collection ID", func() {
				metadata := upload.Metadata{
					CollectionID:  &collectionID,
					FileName:      filename,
					Path:          path,
					IsPublishable: isPublishable,
					Title:         title,
					FileSizeBytes: int64(expectedContentLength),
					FileType:      fileType,
					License:       license,
					LicenseURL:    licenseURL,
				}

				numberOfAPICalls = 0
				err := c.Upload(context.Background(), f, metadata)

				Convey("Then the file is successfully uploaded in 5 Megabyte chunk", func() {
					So(err, ShouldBeNil)
					So(actualMethod, ShouldEqual, http.MethodPost)
					//So(actualContent, ShouldEqual, fileContent)
					So(numberOfAPICalls, ShouldEqual, 2)
				})

				Convey("And the resumable data was calculated", func() {
					So(actualResumableFilename, ShouldEqual, filename)
					So(actualResumableTotalSize, ShouldEqual, fmt.Sprintf("%d", expectedContentLength))
					So(actualResumableChunkNumber, ShouldEqual, "2")
					So(actualResumableTotalChunks, ShouldEqual, "2")
					So(actualResumableType, ShouldEqual, fileType)
				})

				Convey("And the file metadata is sent with the file", func() {
					So(actualCollectionId, ShouldEqual, collectionID)
					So(actualPath, ShouldEqual, path)
					So(actualIsPublishable, ShouldEqual, strconv.FormatBool(isPublishable))
					So(actualTitle, ShouldEqual, title)
					So(actualLicence, ShouldEqual, license)
					So(actualLicenceURL, ShouldEqual, licenseURL)
				})
			})
		})
	})
}

func extractFields(r *http.Request) {
	numberOfAPICalls++
	maxMemory := int64(7 * 1024 * 1024)
	r.ParseMultipartForm(maxMemory)

	actualHasCollectionID = r.Form.Has("collectionId")

	actualCollectionId = r.Form.Get("collectionId")
	actualResumableFilename = r.Form.Get("resumableFilename")
	actualPath = r.Form.Get("path")
	actualIsPublishable = r.Form.Get("isPublishable")
	actualTitle = r.Form.Get("title")
	actualResumableTotalSize = r.Form.Get("resumableTotalSize")
	actualResumableType = r.Form.Get("resumableType")
	actualLicence = r.Form.Get("licence")
	actualLicenceURL = r.Form.Get("licenceURL")
	actualResumableChunkNumber = r.Form.Get("resumableChunkNumber")
	actualResumableTotalChunks = r.Form.Get("resumableTotalChunks")
	actualMethod = r.Method

	contentReader, _, _ := r.FormFile("file")

	contentBytes, _ := io.ReadAll(contentReader)

	actualContent = actualContent + string(contentBytes)
}
