package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
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

var checkResponseBase = func(mockRCHTTPCli *rchttp.ClienterMock, expectedMethod string, expectedUri string) {
	So(len(mockRCHTTPCli.DoCalls()), ShouldEqual, 1)
	So(mockRCHTTPCli.DoCalls()[0].Req.URL.RequestURI(), ShouldEqual, expectedUri)
	So(mockRCHTTPCli.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
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
		check := initialState

		Convey("when datasetClient.Checker is called", func() {
			err := datasetClient.Checker(ctx, &check)
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
		check := initialState

		Convey("when datasetClient.Checker is called", func() {
			err := datasetClient.Checker(ctx, &check)
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
		check := initialState

		Convey("when datasetClient.Checker is called", func() {
			err := datasetClient.Checker(ctx, &check)
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
		check := initialState

		Convey("when datasetClient.Checker is called", func() {
			err := datasetClient.Checker(ctx, &check)
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
		check := initialState

		Convey("when datasetClient.Checker is called", func() {
			err := datasetClient.Checker(ctx, &check)
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

func TestClient_PutVersion(t *testing.T) {

	checkResponse := func(mockRCHTTPCli *rchttp.ClienterMock, expectedVersion Version) {

		checkResponseBase(mockRCHTTPCli, http.MethodPut, "/datasets/123/editions/2017/versions/1")

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

	Convey("given a 200 status with valid empty body is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{}`))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when GetInstance is called", func() {
			instance, err := cli.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, Instance{})
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodGet, "/instances/123")
			})
		})
	})

	Convey("given a 200 status with empty body is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when GetInstance is called", func() {
			_, err := cli.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodGet, "/instances/123")
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
				checkResponseBase(mockRCHTTPCli, http.MethodGet, "/instances/123")
			})
		})
	})
}

func TestClient_GetInstances(t *testing.T) {

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
			_, err := cli.GetInstances(ctx, userAuthToken, serviceAuthToken, collectionID, url.Values{})

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodGet, "/instances")
			})
		})

		Convey("When GetInstance is called with filters", func() {
			_, err := cli.GetInstances(ctx, userAuthToken, serviceAuthToken, collectionID, url.Values{
				"id":      []string{"123"},
				"version": []string{"999"},
			})

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time with the expected query parameters", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodGet, "/instances?id=123&version=999")
			})
		})
	})

}

func Test_PutInstanceImportTasks(t *testing.T) {

	data := InstanceImportTasks{
		ImportObservations: &ImportObservationsTask{State: StateSubmitted.String()},
		BuildHierarchyTasks: []*BuildHierarchyTask{
			&BuildHierarchyTask{DimensionName: "dimension1", State: StateCompleted.String()},
			&BuildHierarchyTask{DimensionName: "dimension2", State: StateCreated.String()},
		},
		BuildSearchIndexTasks: []*BuildSearchIndexTask{
			&BuildSearchIndexTask{State: StateSubmitted.String()},
			&BuildSearchIndexTask{State: StateCompleted.String()},
		},
	}

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"buddy":"ook"}`))),
				}, nil
			},
		}

		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when PutInstanceImportTasks is called", func() {
			err := cli.PutInstanceImportTasks(ctx, serviceAuthToken, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPut, "/instances/123/import_tasks")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_PostInstanceDimensions(t *testing.T) {

	optionsToPost := OptionPost{
		Name:     "testName",
		Option:   "testOption",
		Label:    "testLabel",
		CodeList: "testCodeList",
		Code:     "testCode",
	}

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}
		expectedPayload, err := json.Marshal(optionsToPost)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			err := cli.PostInstanceDimensions(ctx, serviceAuthToken, "123", optionsToPost)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPost, "/instances/123/dimensions")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("wrong!"))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}
		expectedPayload, err := json.Marshal(optionsToPost)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			err := cli.PostInstanceDimensions(ctx, serviceAuthToken, "123", optionsToPost)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123/dimensions, body: wrong!").Error())
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPost, "/instances/123/dimensions")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_PutInstanceState(t *testing.T) {

	data := stateData{
		State: StateCompleted.String(),
	}

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceState is called", func() {
			err := cli.PutInstanceState(ctx, serviceAuthToken, "123", StateCompleted)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

}

func Test_UpdateInstanceWithNewInserts(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}

		Convey("when UpdateInstanceWithNewInserts is called", func() {
			err := cli.UpdateInstanceWithNewInserts(ctx, serviceAuthToken, "123", 999)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPut, "/instances/123/inserted_observations/999")
			})
		})
	})

}

func TestClient_PutInstanceData(t *testing.T) {

	data := JobInstance{
		HeaderNames:          []string{"header1", "header2"},
		NumberOfObservations: 50,
	}

	Convey("given a 200 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			err := cli.PutInstanceData(ctx, serviceAuthToken, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and rchttpclient.Do is called 1 time", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockRCHTTPCli := &rchttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("wrong!"))),
				}, nil
			},
		}

		cli := Client{cli: mockRCHTTPCli, url: "http://localhost:8080"}
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			err := cli.PutInstanceData(ctx, serviceAuthToken, "123", data)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123, body: wrong!").Error())
			})

			Convey("and rchttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockRCHTTPCli, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(mockRCHTTPCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}
