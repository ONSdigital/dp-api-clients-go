package articles

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
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testHost = "http://localhost:8080"
)

func TestGetLegacyBulletin(t *testing.T) {
	accessToken := "token"
	collectionId := "collection"
	lang := "en"
	url := "/gdp/economy"
	expectedArticlesApiUrl := fmt.Sprintf("%s/articles/legacy?url=%s&lang=%s", testHost, url, lang)
	expectedBulletin := Bulletin{
		URI:  url,
		Type: "bulletin",
		Sections: []zebedee.Section{
			{
				Title:    "Section 1",
				Markdown: "Markdown 1",
			},
		},
		LatestReleaseURI: "the/latest/release",
		Links: []zebedee.Link{
			{
				Title: "Link 1",
				URI:   "/link/1",
			},
		},
	}
	bulletinBody, _ := json.Marshal(expectedBulletin)

	Convey("Given that 200 OK is returned by articles API with a valid bulletin body", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(bulletinBody)),
		}, nil)
		articlesClient := newArticlesClient(httpClient)

		Convey("When GetLegacyBulletin is called", func() {
			bulletin, err := articlesClient.GetLegacyBulletin(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the articles API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedArticlesApiUrl)
				So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, http.MethodGet)

				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionId)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, accessToken)
			})
			Convey("And the expected bulletin is returned without error", func() {
				So(err, ShouldBeNil)
				So(*bulletin, ShouldResemble, expectedBulletin)
			})
		})
	})

	Convey("Given that 200 OK is returned by articles API with an invalid body", t, func() {
		responseBody := "invalidBulletin"
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
		}, nil)
		articlesClient := newArticlesClient(httpClient)

		Convey("When GetLegacyBulletin is called", func() {
			bulletin, err := articlesClient.GetLegacyBulletin(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the articles API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedArticlesApiUrl)
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
				So(bulletin, ShouldBeNil)
			})
		})
	})

	Convey("Given that 200 OK is returned by articles API with an empty body", t, func() {
		responseBody := ""
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(responseBody))),
		}, nil)
		articlesClient := newArticlesClient(httpClient)

		Convey("When GetLegacyBulletin is called", func() {
			bulletin, err := articlesClient.GetLegacyBulletin(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the articles API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedArticlesApiUrl)
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
				So(bulletin, ShouldBeNil)
			})
		})
	})

	Convey("Given that 404 is returned by articles API ", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("URL not found"))),
		}, nil)
		articlesClient := newArticlesClient(httpClient)

		Convey("When GetLegacyBulletin is called", func() {
			bulletin, err := articlesClient.GetLegacyBulletin(context.Background(), accessToken, collectionId, lang, url)

			Convey("Then the expected call to the articles API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedArticlesApiUrl)
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
				So(bulletin, ShouldBeNil)
			})
		})
	})

	Convey("Given an http client that fails to perform a request", t, func() {
		errorString := "articles API error"
		httpClient := newMockHTTPClient(nil, errors.New(errorString))
		articlesClient := newArticlesClient(httpClient)

		Convey("When GetLegacyBulletin is called", func() {
			bulletin, err := articlesClient.GetLegacyBulletin(context.Background(), accessToken, collectionId, lang, url)
			Convey("Then the expected call to the articles API is made", func() {
				So(httpClient.DoCalls(), ShouldHaveLength, 1)
				So(httpClient.DoCalls()[0].Req.URL.String(), ShouldEqual, expectedArticlesApiUrl)
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
					errors.New(fmt.Sprintf("failed to get response from Articles API: %s", errorString)),
					http.StatusInternalServerError,
					nil),
				)
				So(bulletin, ShouldBeNil)
			})
		})
	})
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

func newArticlesClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	return NewWithHealthClient(healthClient)
}
