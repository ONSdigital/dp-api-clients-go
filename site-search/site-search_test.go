package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	userAuthToken    = "iamatoken"
	serviceAuthToken = "iamaservicetoken"
	collectionID     = "iamacollectionID"
	testHost         = "http://localhost:8080"
)

var (
	ctx          = context.Background()
	initialState = health.CreateCheckState(service)
)

var checkResponseBase = func(mockdphttpCli *dphttp.ClienterMock, expectedMethod string, expectedUri string) {
	So(len(mockdphttpCli.DoCalls()), ShouldEqual, 1)
	So(mockdphttpCli.DoCalls()[0].Req.URL.RequestURI(), ShouldEqual, expectedUri)
	So(mockdphttpCli.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := createHTTPClientMockErr(clientError)
		searchClient := newSearchClient(httpClient)
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 400 response", t, func() {
		httpClient := createHTTPClientMock(http.StatusBadRequest, []byte(""))
		searchClient := newSearchClient(httpClient)
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 400)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		httpClient := createHTTPClientMock(http.StatusInternalServerError, []byte(""))
		searchClient := newSearchClient(httpClient)
		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_GetSearch(t *testing.T) {
	Convey("given a 200 status is returned with an empty result list", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/empty_results.json")
		So(err, ShouldBeNil)

		httpClient := createHTTPClientMock(http.StatusOK, searchResp)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("q", "a")
			r, err := searchClient.GetSearch(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 0)
				So(r.ContentTypes, ShouldBeEmpty)
				So(r.Items, ShouldBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search?q=a")
			})
		})
	})

	Convey("given a 200 status is returned with list of search results", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/results.json")
		So(err, ShouldBeNil)

		httpClient := createHTTPClientMock(http.StatusOK, searchResp)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("q", "housing")
			r, err := searchClient.GetSearch(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 5)
				So(r.Items, ShouldNotBeEmpty)
				So(r.ContentTypes, ShouldNotBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search?q=housing")
			})
		})
	})

	Convey("given a 400 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusBadRequest, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "a")
			_, err := searchClient.GetSearch(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 400, path: "+testHost+"/search?limit=a").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search?limit=a")
			})
		})
	})

	Convey("given a 500 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusInternalServerError, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "housing")
			_, err := searchClient.GetSearch(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 500, path: "+testHost+"/search?limit=housing").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search?limit=housing")
			})
		})
	})
}

func TestClient_GetDepartments(t *testing.T) {
	Convey("given a 200 status is returned with an empty result list", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/empty_results.json")
		So(err, ShouldBeNil)

		httpClient := createHTTPClientMock(http.StatusOK, searchResp)
		searchClient := newSearchClient(httpClient)

		Convey("when GetDepartments is called", func() {
			v := url.Values{}
			v.Set("q", "a")
			r, err := searchClient.GetDepartments(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 0)
				So(r.Items, ShouldBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/departments/search?q=a")
			})
		})
	})

	Convey("given a 200 status is returned with list of department search results", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/departments.json")
		So(err, ShouldBeNil)

		httpClient := createHTTPClientMock(http.StatusOK, searchResp)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("q", "housing")
			r, err := searchClient.GetDepartments(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(r.Count, ShouldEqual, 1)
				So(r.Items, ShouldNotBeEmpty)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/departments/search?q=housing")
			})
		})
	})

	Convey("given a 400 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusBadRequest, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "a")
			_, err := searchClient.GetDepartments(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 400, path: "+testHost+"/departments/search?limit=a").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/departments/search?limit=a")
			})
		})
	})

	Convey("given a 500 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusInternalServerError, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "housing")
			_, err := searchClient.GetDepartments(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 500, path: "+testHost+"/departments/search?limit=housing").Error())

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(httpClient, http.MethodGet, "/departments/search?limit=housing")
			})
		})
	})
}

func TestClient_GetReleases(t *testing.T) {
	releaseResponse := ReleaseResponse{
		Took: 100,
		Breakdown: Breakdown{
			Total: 11,
		},
		Releases: []Release{
			{
				URI: "/releases/title1",
				Description: ReleaseDescription{
					Title:           "Public Sector Employment, UK: September 2021",
					Summary:         "A summary for Title 1",
					ReleaseDate:     time.Now().AddDate(0, 0, -10).UTC().Format(time.RFC3339),
					Published:       true,
					Finalised:       true,
					Census:          true,
					Keywords:        []string{"2020", "census"},
					ProvisionalDate: "2020-01-12",
					Language:        "en",
				},
			},
			{
				URI: "/releases/title2",
				DateChanges: []ReleaseDateChange{
					{
						Date:         "2015-09-22T12:30:23.221Z",
						ChangeNotice: "Something happened to change the date",
					},
				},
				Description: ReleaseDescription{
					Title:       "Labour market in the regions of the UK: December 2021",
					Summary:     "A summary for Title 2",
					ReleaseDate: time.Now().AddDate(0, 0, -15).UTC().Format(time.RFC3339),
					Published:   true,
					Finalised:   true,
					Postponed:   true,
					Keywords:    []string{"something"},
					Language:    "cy",
				},
			},
		},
	}
	releaseResponseBody, _ := json.Marshal(releaseResponse)

	Convey("given a 200 status is returned with a list of release calendar entries", t, func() {
		httpClient := createHTTPClientMock(http.StatusOK, releaseResponseBody)
		searchClient := newSearchClient(httpClient)

		Convey("when GetReleases is called", func() {
			v := url.Values{}
			v.Set("limit", "1")
			v.Set("q", "answer")
			rr, err := searchClient.GetReleases(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("the expected call to the search API is made", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search/releases?limit=1&q=answer")
				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionID)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, userAuthToken)

			})

			Convey("and the expected calendar is returned without error", func() {
				So(err, ShouldBeNil)
				So(rr, ShouldResemble, releaseResponse)
			})
		})
	})

	Convey("Given that 200 OK is returned by the API with an invalid body", t, func() {
		responseBody := "invalidRelease"
		httpClient := createHTTPClientMock(http.StatusOK, []byte(responseBody))
		searchClient := newSearchClient(httpClient)

		Convey("when GetReleases is called", func() {
			v := url.Values{}
			v.Set("limit", "1")
			v.Set("q", "answer")
			rr, err := searchClient.GetReleases(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("the expected call to the search API is made", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search/releases?limit=1&q=answer")
				collectionHeader, err := headers.GetCollectionID(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(collectionHeader, ShouldEqual, collectionID)

				authTokenHeader, err := headers.GetUserAuthToken(httpClient.DoCalls()[0].Req)
				So(err, ShouldBeNil)
				So(authTokenHeader, ShouldEqual, userAuthToken)

			})

			Convey("And an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(rr, ShouldBeZeroValue)
			})
		})
	})

	Convey("given a 400 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusBadRequest, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "1")
			v.Set("q", "answer")
			_, err := searchClient.GetReleases(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 400, path: "+testHost+"/search/releases?limit=1&q=answer").Error())

			})

			Convey("and dphttpclient.Do is called once", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search/releases?limit=1&q=answer")
			})
		})
	})

	Convey("given a 500 status is returned", t, func() {
		httpClient := createHTTPClientMock(http.StatusInternalServerError, nil)
		searchClient := newSearchClient(httpClient)

		Convey("when GetSearch is called", func() {
			v := url.Values{}
			v.Set("limit", "1")
			v.Set("q", "answer")
			_, err := searchClient.GetReleases(ctx, userAuthToken, serviceAuthToken, collectionID, v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, fmt.Errorf("invalid response from dp-search-api - should be: 200, got: 500, path: "+testHost+"/search/releases?limit=1&q=answer").Error())

			})

			Convey("and dphttpclient.Do is called once", func() {
				checkResponseBase(httpClient, http.MethodGet, "/search/releases?limit=1&q=answer")
			})
		})
	})
}

func newSearchClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter(service, testHost, httpClient)
	searchClient := NewWithHealthClient(healthClient)
	return searchClient
}

func createHTTPClientMock(retCode int, body []byte) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: retCode,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func createHTTPClientMockErr(err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return nil, err
		},
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}
