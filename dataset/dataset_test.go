package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
)

var ctx = context.Background()

const (
	userAuthToken    = "iamatoken"
	serviceAuthToken = "iamaservicetoken"
	collectionID     = "iamacollectionID"
	testHost         = "http://localhost:8080"
)

var checkResponseBase = func(mockRCHTTPCli *rchttp.ClienterMock) {
	So(len(mockRCHTTPCli.DoCalls()), ShouldEqual, 1)
	So(mockRCHTTPCli.DoCalls()[0].Req.Header.Get(common.AuthHeaderKey), ShouldEqual, "Bearer "+serviceAuthToken)
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		datasetClient := NewAPIClient(testHost)
		datasetClient.cli = clienter

		Convey("when datasetClient.Checker is called", func() {
			check, err := datasetClient.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 0)
				So(check.Message, ShouldEqual, clientError.Error())
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		datasetClient := NewAPIClient(testHost)
		datasetClient.cli = clienter

		Convey("when datasetClient.Checker is called", func() {
			check, err := datasetClient.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 500)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		datasetClient := NewAPIClient(testHost)
		datasetClient.cli = clienter

		Convey("when datasetClient.Checker is called", func() {
			check, err := datasetClient.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode, ShouldEqual, 404)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 429,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		datasetClient := NewAPIClient(testHost)
		datasetClient.cli = clienter

		Convey("when datasetClient.Checker is called", func() {
			check, err := datasetClient.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode, ShouldEqual, 429)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess, ShouldBeNil)
				So(*check.LastFailure, ShouldHappenAfter, timePriorHealthCheck)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		clienter := &rchttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		datasetClient := NewAPIClient(testHost)
		datasetClient.cli = clienter

		Convey("when datasetClient.Checker is called", func() {
			check, err := datasetClient.Checker(ctx)

			Convey("then the expected check is returned", func() {
				So(check.Name, ShouldEqual, service)
				So(check.Status, ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode, ShouldEqual, 200)
				So(check.Message, ShouldEqual, health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked, ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess, ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure, ShouldBeNil)
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_PutVersion(t *testing.T) {

	checkResponse := func(mockRCHTTPCli *rchttp.ClienterMock, expectedVersion Version) {

		checkResponseBase(mockRCHTTPCli)

		actualBody, _ := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)

		var actualVersion Version
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedVersion)
	}

	Convey("Given a valid version", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

	Convey("Given no auth token has been configured", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttp client is called one time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})

		})
	})

	Convey("given rchttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular explosion")
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return nil, mockErr
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

	Convey("given rchttpclient.do returns a non 200 response status", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := cli.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("incorrect http status, expected: 200, actual: 500, uri: http://localhost:8080/datasets/123/editions/2017/versions/1").Error())
			})

			Convey("and rchttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(mockRCHTTPCli, v)
			})
		})
	})

}

func TestClient_IncludeCollectionID(t *testing.T) {

	Convey("Given a valid request", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{Body: ioutil.NopCloser(nil)}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Collection-ID is present in the context", func() {
			ctx = context.WithValue(ctx, common.CollectionIDHeaderKey, collectionID)

			Convey("and a request is made", func() {
				_, _ = cli.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID)

				Convey("then the Collection-ID is present in the request headers", func() {
					collectionIDFromRequest := mockRCHTTPCli.DoCalls()[0].Req.Header.Get(common.CollectionIDHeaderKey)
					So(collectionIDFromRequest, ShouldEqual, collectionID)
				})
			})
		})
	})

	Convey("Given a valid request", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{Body: ioutil.NopCloser(nil)}, nil
			},
		}

		cli := Client{
			cli: mockRCHTTPCli,
			url: "http://localhost:8080",
		}

		Convey("when Collection-ID is not present in the context", func() {
			ctx = context.Background()

			Convey("and a request is made", func() {
				_, _ = cli.GetDatasets(ctx, userAuthToken, serviceAuthToken, "")

				Convey("then no Collection-ID key is present in the request headers", func() {
					for k := range mockRCHTTPCli.DoCalls()[0].Req.Header {
						So(k, ShouldNotEqual, "Collection-Id")
					}
				})
			})
		})
	})
}

func TestClient_GetInstance(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"buddy":"ook"}`))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when GetInstance is called", func() {
			_, err := cli.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("you aint seen me roight"))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when GetInstance is called", func() {
			_, err := cli.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123, body: you aint seen me roight").Error())
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli)
			})
		})
	})
}
