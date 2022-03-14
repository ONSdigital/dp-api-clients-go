package upload_test

import (
	"context"
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
	actualContent, actualCollectionId string
	actualResumableFilename           string
	actualPath                        string
	actualIsPublishable               string
	actualTitle                       string
	actualResumableTotalSize          string
	actualResumableType               string
	actualLicence                     string
	actualLicenceURL                  string
	actualResumableChunkNumber        string
	actualResumableTotalChunks        string
	actualMethod                      string
)

const (
	collectionID  = "123456"
	filename      = "file.txt"
	path          = "data/file.txt"
	isPublishable = false
	title         = "Information about shoe size"
	fileType      = "text/plain"
	license       = "MIT"
	licenseURL    = "https://opensource.org/licenses/MIT"
)

func TestHealthCheck(t *testing.T) {
	Convey("Given the upload service is health", t, func() {
		timePriorHealthCheck := time.Now()
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		defer s.Close()

		state := health.CreateCheckState("testing")

		Convey("When we check that state of the service", func() {
			c := upload.NewAPIClient(s.URL)
			c.Checker(context.Background(), &state)

			So(state.Status(), ShouldEqual, healthcheck.StatusOK)
			So(state.StatusCode(), ShouldEqual, 200)
			So(state.Message(), ShouldContainSubstring, "is ok")

			So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
			So(*state.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
			So(state.LastFailure(), ShouldBeNil)
		})
	})

	Convey("Given the upload service is failing", t, func() {
		timePriorHealthCheck := time.Now()
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
		defer s.Close()

		state := health.CreateCheckState("testing")

		Convey("When we check the state of the service", func() {
			c := upload.NewAPIClient(s.URL)
			c.Checker(context.Background(), &state)

			So(state.Status(), ShouldEqual, healthcheck.StatusCritical)
			So(state.StatusCode(), ShouldEqual, 500)
			So(state.Message(), ShouldContainSubstring, "unavailable or non-functioning")

			So(*state.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
			So(state.LastSuccess(), ShouldBeNil)
			So(*state.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
		})
	})
}

func TestUpload(t *testing.T) {
	Convey("Given the client uploads a single chunk file", t, func() {
		fileContent := "testing"
		f := io.NopCloser(strings.NewReader(fileContent))

		Convey("And the file belongs to a collection", func() {

			metadata := upload.Metadata{
				CollectionID:  collectionID,
				FileName:      filename,
				Path:          path,
				IsPublishable: isPublishable,
				Title:         title,
				FileSizeBytes: int64(len(fileContent)),
				FileType:      fileType,
				License:       license,
				LicenseURL:    licenseURL,
			}

			Convey("When the upload is successful", func() {
				ctx := context.Background()

				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					extractFields(r)

					w.WriteHeader(http.StatusCreated)
				},
				))
				c := upload.NewAPIClient(s.URL)

				err := c.Upload(ctx, f, metadata)

				So(err, ShouldBeNil)
				So(actualMethod, ShouldEqual, http.MethodPost)

				So(actualCollectionId, ShouldEqual, collectionID)
				So(actualResumableFilename, ShouldEqual, filename)
				So(actualPath, ShouldEqual, path)
				So(actualIsPublishable, ShouldEqual, strconv.FormatBool(isPublishable))
				So(actualTitle, ShouldEqual, title)
				So(actualResumableTotalSize, ShouldEqual, fmt.Sprintf("%d", len(fileContent)))
				So(actualResumableType, ShouldEqual, fileType)
				So(actualLicence, ShouldEqual, license)
				So(actualLicenceURL, ShouldEqual, licenseURL)
				So(actualResumableChunkNumber, ShouldEqual, "1")
				So(actualResumableTotalChunks, ShouldEqual, "1")

				So(actualContent, ShouldEqual, fileContent)
			})
		})

	})
}

func extractFields(r *http.Request) {
	r.ParseMultipartForm(4)

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

	content, _, _ := r.FormFile("file")
	by, _ := io.ReadAll(content)
	actualContent = string(by)
}
