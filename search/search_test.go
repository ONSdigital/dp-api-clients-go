package search

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/golang/mock/gomock"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx                  = context.Background()
	defaultLimitAsString = strconv.Itoa(defaultLimit)
	initialState         = health.CreateCheckState(service)
)

const (
	clientErrText = "client threw an error"
	testHost      = "http://localhost:8080"
)

func TestSearchUnit(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	limit := 1
	offset := 1

	Convey("test New creates a valid Client instance", t, func() {
		cli := New("http://localhost:22000")
		So(cli.hcCli.URL, ShouldEqual, "http://localhost:22000")
		So(cli.hcCli.Client, ShouldHaveSameTypeAs, dphttp.DefaultClient)
	})

	Convey("test Dimension Method", t, func() {
		searchResp, err := ioutil.ReadFile("./search_mocks/search.json")
		So(err, ShouldBeNil)

		Convey("test Dimension successfully returns a model upon a 200 response from search api", func() {

			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(searchResp))),
					}, nil
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			ctx := context.Background()

			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err, ShouldBeNil)
			So(m.Count, ShouldEqual, 1)
			So(m.Limit, ShouldEqual, 1)
			So(m.Offset, ShouldEqual, 0)
			So(m.TotalCount, ShouldEqual, 1)
			So(m.Items, ShouldHaveLength, 1)

			item := m.Items[0]
			So(item.Code, ShouldEqual, "6789")
			So(item.DimensionOptionURL, ShouldEqual, "http://localhost:22000/datasets/12345/editions/time-series/versions/1/dimensions/geography/options/6789")
			So(item.HasData, ShouldBeTrue)
			So(item.Label, ShouldEqual, "Newport")
			So(item.Matches.Label, ShouldHaveLength, 1)
			So(item.NumberOfChildren, ShouldEqual, 3)

			label := item.Matches.Label[0]
			So(label.Start, ShouldEqual, 0)
			So(label.End, ShouldEqual, 6)
		})

		Convey("test Dimension returns error from HTTPClient if it throws an error", func() {
			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New(clientErrText)
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err.Error(), ShouldEqual, clientErrText)
			So(m, ShouldBeNil)
		})

		Convey("test Dimension returns error if HTTP Status code is not 200", func() {

			searchErr := errors.New("invalid response from search api - should be: 200, got: 400, path: http://localhost:22000/dimension-search/datasets/12345/editions/time-series/versions/1/dimensions/geography?limit=1&offset=1&q=Newport")
			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, searchErr
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Limit: &limit, Offset: &offset})
			So(err, ShouldEqual, searchErr)
			So(m, ShouldBeNil)
		})

		Convey("test Dimension uses default search limit when no limit config provided", func() {

			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(searchResp))),
					}, nil
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			Convey("when dimension search is called", func() {
				m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Offset: &offset})

				Convey("then the request is sent with the default limit", func() {
					So(mockClient.DoCalls(), ShouldHaveLength, 1)
					q := mockClient.DoCalls()[0].Req.URL.Query()
					So(q.Get("limit"), ShouldEqual, defaultLimitAsString)
				})

				Convey("and the expected model is returned", func() {
					So(err, ShouldBeNil)
					So(m.Count, ShouldEqual, 1)
					So(m.Limit, ShouldEqual, 1)
					So(m.Offset, ShouldEqual, 0)
					So(m.TotalCount, ShouldEqual, 1)
					So(m.Items, ShouldHaveLength, 1)

					item := m.Items[0]
					So(item.Code, ShouldEqual, "6789")
					So(item.DimensionOptionURL, ShouldEqual, "http://localhost:22000/datasets/12345/editions/time-series/versions/1/dimensions/geography/options/6789")
					So(item.HasData, ShouldBeTrue)
					So(item.Label, ShouldEqual, "Newport")
					So(item.Matches.Label, ShouldHaveLength, 1)
					So(item.NumberOfChildren, ShouldEqual, 3)

					label := item.Matches.Label[0]
					So(label.Start, ShouldEqual, 0)
					So(label.End, ShouldEqual, 6)
				})
			})
		})

		Convey("test Dimension no limit returns error from HTTPClient if it throws an error", func() {

			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return nil, errors.New(clientErrText)
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			Convey("when search is called", func() {
				m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Offset: &offset})

				Convey("then the request is sent with the default limit", func() {
					So(mockClient.DoCalls(), ShouldHaveLength, 1)
					q := mockClient.DoCalls()[0].Req.URL.Query()
					So(q.Get("limit"), ShouldEqual, defaultLimitAsString)
				})

				Convey("and the expected error is returned", func() {
					So(err.Error(), ShouldEqual, clientErrText)
					So(m, ShouldBeNil)
				})

			})
		})

		Convey("test Dimension no limit returns error if HTTP Status code is not 200", func() {

			expectedError := &ErrInvalidSearchAPIResponse{http.StatusOK, http.StatusTeapot, "http://localhost:22000/dimension-search/datasets/12345/editions/time-series/versions/1/dimensions/geography?limit=50&offset=1&q=Newport"}
			mockClient := &dphttp.ClienterMock{
				GetPathsWithNoRetriesFunc: func() []string { return []string{} },
				SetPathsWithNoRetriesFunc: func([]string) {},
				DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusTeapot,
						Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
					}, nil
				},
			}

			hcCli := health.NewClientWithClienter(service, "http://localhost:22000", mockClient)
			searchCli := NewWithHealthClient(hcCli)

			Convey("when search is called", func() {
				m, err := searchCli.Dimension(ctx, "12345", "time-series", "1", "geography", "Newport", Config{Offset: &offset})

				Convey("then the request is sent with the default limit", func() {
					So(mockClient.DoCalls(), ShouldHaveLength, 1)
					q := mockClient.DoCalls()[0].Req.URL.Query()
					So(q.Get("limit"), ShouldEqual, defaultLimitAsString)
				})

				Convey("and the expected error is returned", func() {
					So(err, ShouldResemble, expectedError)
					So(m, ShouldBeNil)
				})

			})

		})
	})
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &dphttp.ClienterMock{
			GetPathsWithNoRetriesFunc: func() []string { return []string{} },
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		hcCli := health.NewClientWithClienter(service, testHost, clienter)
		searchClient := NewWithHealthClient(hcCli)

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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		clienter := &dphttp.ClienterMock{
			GetPathsWithNoRetriesFunc: func() []string { return []string{} },
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

		hcCli := health.NewClientWithClienter(service, testHost, clienter)
		searchClient := NewWithHealthClient(hcCli)

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
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		clienter := &dphttp.ClienterMock{
			GetPathsWithNoRetriesFunc: func() []string { return []string{} },
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

		hcCli := health.NewClientWithClienter(service, testHost, clienter)
		searchClient := NewWithHealthClient(hcCli)

		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 404)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
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
		clienter := &dphttp.ClienterMock{
			GetPathsWithNoRetriesFunc: func() []string { return []string{} },
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

		hcCli := health.NewClientWithClienter(service, testHost, clienter)
		searchClient := NewWithHealthClient(hcCli)

		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, 429)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		clienter := &dphttp.ClienterMock{
			GetPathsWithNoRetriesFunc: func() []string { return []string{} },
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

		hcCli := health.NewClientWithClienter(service, testHost, clienter)
		searchClient := NewWithHealthClient(hcCli)

		check := initialState

		Convey("when searchClient.Checker is called", func() {
			err := searchClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure(), ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := clienter.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}
