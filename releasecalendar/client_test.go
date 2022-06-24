package releasecalendar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	dphttp "github.com/ONSdigital/dp-net/v2/http"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testHost = "http://localhost:8080"
)

func TestClientNew(t *testing.T) {
	Convey("NewAPIClient creates a new API client with the expected URL and name", t, func() {
		client := NewAPIClient(testHost)
		So(client.URL(), ShouldEqual, testHost)
		So(client.HealthClient().Name, ShouldEqual, "release-calendar-api")
	})

	Convey("Given an existing healthcheck client", t, func() {
		hcClient := health.NewClient("generic", testHost)
		Convey("When creating a new release calendar API client providing it", func() {
			client := NewWithHealthClient(hcClient)
			Convey("Then it returns a new client with the expected URL and name", func() {
				So(client.URL(), ShouldEqual, testHost)
				So(client.HealthClient().Name, ShouldEqual, "release-calendar-api")
			})
		})
	})
}

func TestGetLegacyRelease(t *testing.T) {
	accessToken := "token"
	collectionId := "collection"
	lang := "en"
	url := "/releases/gdpukregionsandcountriesapriltojune2021"
	expectedReleaseCalendarApiUrl := fmt.Sprintf("%s/releases/legacy?url=%s&lang=%s", testHost, url, lang)
	expectedRelease := Release{
		URI:      url,
		Markdown: []string{"markdown1", "markdown 2"},
		RelatedDocuments: []Link{
			{
				Title:   "Document 1",
				Summary: "This is document 1",
				URI:     "/doc/1",
			},
		},
		RelatedDatasets: []Link{
			{
				Title:   "Dataset 1",
				Summary: "This is dataset 1",
				URI:     "/dataset/1",
			},
		},
		RelatedMethodology: []Link{
			{
				Title:   "Methodology",
				Summary: "This is methodology 1",
				URI:     "/methodology/1",
			},
		},
		RelatedMethodologyArticle: []Link{
			{
				Title:   "Methodology Article",
				Summary: "This is methodology article 1",
				URI:     "/methodology/article/1",
			},
		},
		Links: []Link{
			{
				Title:   "Link 1",
				Summary: "This is link 1",
				URI:     "/link/1",
			},
		},
		DateChanges: []ReleaseDateChange{
			{
				Date:         "2022-02-15T11:12:05.592Z",
				ChangeNotice: "This release has changed",
			},
		},
		Description: ReleaseDescription{
			Title:   "Release title",
			Summary: "Release summary",
			Contact: Contact{
				Email:     "contact@ons.gov.uk",
				Name:      "Contact name",
				Telephone: "029",
			},
			NationalStatistic:  true,
			ReleaseDate:        "2020-07-08T23:00:00.000Z",
			NextRelease:        "January 2021",
			Published:          true,
			Finalised:          true,
			Cancelled:          true,
			CancellationNotice: []string{"cancelled for a reason"},
			ProvisionalDate:    "July 2020",
		},
	}
	releaseBody, _ := json.Marshal(expectedRelease)

	Convey("Given that 200 OK is returned by the API with a valid release body", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(releaseBody)),
		}, nil)
		client := newReleaseCalendarApiClient(httpClient)

		Convey("When GetLegacyRelease is called", func() {
			release, err := client.GetLegacyRelease(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the release calendar API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedReleaseCalendarApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And the expected release is returned without error", func() {
				So(err, ShouldBeNil)
				So(*release, ShouldResemble, expectedRelease)
			})
		})
	})

	Convey("Given that 200 OK is returned by the API with an invalid body", t, func() {
		responseBody := "invalidRelease"
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
		}, nil)
		client := newReleaseCalendarApiClient(httpClient)

		Convey("When GetLegacyRelease is called", func() {
			release, err := client.GetLegacyRelease(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the release calendar API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedReleaseCalendarApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And an error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New("failed to unmarshal response body: invalid character 'i' looking for beginning of value"),
					http.StatusInternalServerError,
					log.Data{"response_body": responseBody}),
				)
				So(release, ShouldBeNil)
			})
		})
	})

	Convey("Given that 200 OK is returned by the API with an empty body", t, func() {
		responseBody := ""
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
		}, nil)
		client := newReleaseCalendarApiClient(httpClient)

		Convey("When GetLegacyRelease is called", func() {
			release, err := client.GetLegacyRelease(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the release calendar API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedReleaseCalendarApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And an error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New("failed to unmarshal response body: unexpected end of JSON input"),
					http.StatusInternalServerError,
					log.Data{"response_body": responseBody}),
				)
				So(release, ShouldBeNil)
			})
		})
	})

	Convey("Given that 404 is returned by the API ", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("URL not found"))),
		}, nil)
		client := newReleaseCalendarApiClient(httpClient)

		Convey("When GetLegacyRelease is called", func() {
			release, err := client.GetLegacyRelease(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the release calendar API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedReleaseCalendarApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And an error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New("URL not found"),
					http.StatusNotFound,
					nil),
				)
				So(release, ShouldBeNil)
			})
		})
	})

	Convey("Given an http client that fails to perform a request", t, func() {
		errorString := "release calendar API error"
		httpClient := newMockHTTPClient(nil, errors.New(errorString))
		client := newReleaseCalendarApiClient(httpClient)

		Convey("When GetLegacyRelease is called", func() {
			release, err := client.GetLegacyRelease(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the release calendar API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedReleaseCalendarApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And an error is returned", func() {
				So(err, ShouldResemble, dperrors.New(
					errors.New(fmt.Sprintf("failed to get response from Release Calendar API: %s", errorString)),
					http.StatusInternalServerError,
					nil),
				)
				So(release, ShouldBeNil)
			})
		})
	})

}

func newReleaseCalendarApiClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	return NewWithHealthClient(healthClient)
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
	}
}
