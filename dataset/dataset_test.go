package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
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

var checkRequestBase = func(httpClient *dphttp.ClienterMock, expectedMethod string, expectedUri string) {
	So(len(httpClient.DoCalls()), ShouldEqual, 1)
	So(httpClient.DoCalls()[0].Req.URL.RequestURI(), ShouldResemble, expectedUri)
	So(httpClient.DoCalls()[0].Req.Method, ShouldEqual, expectedMethod)
	So(httpClient.DoCalls()[0].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+serviceAuthToken)
}

// getRequestPatchBody returns the patch request body sent with the provided httpClient in iteration callIndex
var getRequestPatchBody = func(httpClient *dphttp.ClienterMock, callIndex int) []dprequest.Patch {
	sentPayload, err := ioutil.ReadAll(httpClient.DoCalls()[callIndex].Req.Body)
	So(err, ShouldBeNil)
	var sentBody []dprequest.Patch
	err = json.Unmarshal(sentPayload, &sentBody)
	So(err, ShouldBeNil)
	return sentBody
}

type MockedHTTPResponse struct {
	StatusCode int
	Body       interface{}
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := createHTTPClientMockErr(clientError)
		datasetClient := newDatasetClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusInternalServerError, ""})
		datasetClient := newDatasetClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusNotFound, ""},
			MockedHTTPResponse{http.StatusNotFound, ""})
		datasetClient := newDatasetClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusTooManyRequests, ""})
		datasetClient := newDatasetClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)
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
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_PutVersion(t *testing.T) {

	checkResponse := func(httpClient *dphttp.ClienterMock, expectedVersion Version) {

		checkRequestBase(httpClient, http.MethodPut, "/datasets/123/editions/2017/versions/1")

		actualBody, _ := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)

		var actualVersion Version
		json.Unmarshal(actualBody, &actualVersion)
		So(actualVersion, ShouldResemble, expectedVersion)
	}

	Convey("Given a valid version", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := datasetClient.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})
		})
	})

	Convey("Given no auth token has been configured", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := datasetClient.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttp client is called one time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})

		})
	})

	Convey("given dphttpclient.do returns an error", t, func() {
		mockErr := errors.New("spectacular explosion")
		httpClient := createHTTPClientMockErr(mockErr)
		datasetClient := newDatasetClient(httpClient)

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := datasetClient.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Wrap(mockErr, "http client returned error while attempting to make request").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})
		})
	})

	Convey("given dphttpclient.do returns a non 200 response status", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusInternalServerError, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when put version is called", func() {
			v := Version{ID: "666"}
			err := datasetClient.PutVersion(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "2017", "1", v)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("incorrect http status, expected: 200, actual: 500, uri: http://localhost:8080/datasets/123/editions/2017/versions/1").Error())
			})

			Convey("and dphttpclient.do is called 1 time with the expected parameters", func() {
				checkResponse(httpClient, v)
			})
		})
	})

}

func TestClient_GetDatasets(t *testing.T) {

	offset := 1
	limit := 10

	Convey("given a 200 status is returned", t, func() {
		expectedDatasets := List{
			Items: []Dataset{
				{ID: "datasetID1"},
				{ID: "datasetID2"},
			},
			Count:      2,
			Offset:     offset,
			Limit:      limit,
			TotalCount: 3,
		}
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, expectedDatasets})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetDatasets is called with valid values for limit and offset", func() {
			q := QueryParams{Offset: offset, Limit: limit, IDs: []string{}}
			actualDatasets, err := datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID, &q)

			Convey("a positive response is returned, with the expected datasets", func() {
				So(err, ShouldBeNil)
				So(actualDatasets, ShouldResemble, expectedDatasets)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := fmt.Sprintf("/datasets?offset=%d&limit=%d", offset, limit)
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})

		Convey("when GetDatasets is called with negative offset", func() {
			q := QueryParams{Offset: -1, Limit: limit, IDs: []string{}}
			options, err := datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, List{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("when GetDatasets is called with negative limit", func() {
			q := QueryParams{Offset: offset, Limit: -1, IDs: []string{}}
			options, err := datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, List{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, List{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetDatasets is called", func() {
			options, err := datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID, nil)

			Convey("the expected error response is returned, with an empty options struct", func() {
				So(err, ShouldResemble, &ErrInvalidDatasetAPIResponse{
					actualCode: 404,
					uri:        fmt.Sprintf("http://localhost:8080/datasets"),
					body:       "{\"items\":null,\"count\":0,\"offset\":0,\"limit\":0,\"total_count\":0}",
				})
				So(options, ShouldResemble, List{})
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := fmt.Sprintf("/datasets")
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})
	})

	Convey("Given a valid request", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when Collection-ID is present in the context", func() {
			ctx = context.WithValue(ctx, dprequest.CollectionIDHeaderKey, collectionID)

			Convey("and a request is made", func() {
				_, _ = datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, collectionID, nil)

				Convey("then the Collection-ID is present in the request headers", func() {
					collectionIDFromRequest := httpClient.DoCalls()[0].Req.Header.Get(dprequest.CollectionIDHeaderKey)
					So(collectionIDFromRequest, ShouldEqual, collectionID)
				})
			})
		})
	})

	Convey("Given a valid request", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when Collection-ID is not present in the context", func() {
			ctx = context.Background()

			Convey("and a request is made", func() {
				_, _ = datasetClient.GetDatasets(ctx, userAuthToken, serviceAuthToken, "", nil)

				Convey("then no Collection-ID key is present in the request headers", func() {
					for k := range httpClient.DoCalls()[0].Req.Header {
						So(k, ShouldNotEqual, "Collection-Id")
					}
				})
			})
		})
	})
}

func TestClient_GetDatasetsInBatches(t *testing.T) {

	datasetsResponse1 := List{
		Items:      []Dataset{{ID: "testDataset1"}},
		TotalCount: 2, // Total count is read from the first response to determine how many batches are required
		Offset:     0,
		Count:      1,
	}

	datasetsResponse2 := List{
		Items:      []Dataset{{ID: "testDataset2"}},
		TotalCount: 2,
		Offset:     1,
		Count:      1,
	}

	expectedDatasets := List{
		Items: []Dataset{
			datasetsResponse1.Items[0],
			datasetsResponse2.Items[0],
		},
		Count:      2,
		TotalCount: 2,
	}

	batchSize := 1
	maxWorkers := 1

	Convey("When a 200 OK status is returned in 2 consecutive calls", t, func() {

		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusOK, datasetsResponse1},
			MockedHTTPResponse{http.StatusOK, datasetsResponse2})
		datasetClient := newDatasetClient(httpClient)

		processedBatches := []List{}
		var testProcess DatasetsBatchProcessor = func(batch List) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetDatasetsInBatches succeeds and returns the accumulated items from all the batches", func() {
			datasets, err := datasetClient.GetDatasetsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, batchSize, maxWorkers)

			So(err, ShouldBeNil)
			So(datasets, ShouldResemble, expectedDatasets)
		})

		Convey("then GetDatasetsBatchProcess calls the batchProcessor function twice, with the expected batches", func() {
			err := datasetClient.GetDatasetsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, testProcess, batchSize, maxWorkers)
			So(err, ShouldBeNil)
			So(processedBatches, ShouldResemble, []List{datasetsResponse1, datasetsResponse2})
			So(httpClient.DoCalls(), ShouldHaveLength, 2)
			So(httpClient.DoCalls()[0].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets?offset=0&limit=1")
			So(httpClient.DoCalls()[1].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets?offset=1&limit=1")
		})
	})

	Convey("When a 400 error status is returned in the first call", t, func() {
		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusBadRequest, ""})
		datasetClient := newDatasetClient(httpClient)

		processedBatches := []List{}
		var testProcess DatasetsBatchProcessor = func(batch List) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetOptionsInBatches fails with the expected error and the process is aborted", func() {
			_, err := datasetClient.GetDatasetsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets?offset=0&limit=1")
		})

		Convey("then GetDatasetsBatchProcess fails with the expected error and doesn't call the batchProcessor", func() {
			err := datasetClient.GetDatasetsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, testProcess, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets?offset=0&limit=1")
			So(processedBatches, ShouldResemble, []List{})
		})
	})

	Convey("When a 200 error status is returned in the first call and 400 error is returned in the second call", t, func() {
		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusOK, datasetsResponse1},
			MockedHTTPResponse{http.StatusBadRequest, ""})
		datasetClient := newDatasetClient(httpClient)

		// testProcess is a generic batch processor for testing
		processedBatches := []List{}
		var testProcess DatasetsBatchProcessor = func(batch List) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetDatasetsInBatches fails with the expected error, corresponding to the second batch, and the process is aborted", func() {
			_, err := datasetClient.GetDatasetsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets?offset=1&limit=1")
		})

		Convey("then GetDatasetsBatchProcess fails with the expected error and calls the batchProcessor for the first batch only", func() {
			err := datasetClient.GetDatasetsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, testProcess, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets?offset=1&limit=1")
			So(processedBatches, ShouldResemble, []List{datasetsResponse1})
		})
	})

}

func TestClient_GetDatasetCurrentAndNext(t *testing.T) {

	Convey("given a 200 status with valid empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, Dataset{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetDatasetCurrentAndNext is called", func() {
			instance, err := datasetClient.GetDatasetCurrentAndNext(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, Dataset{})
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/datasets/123")
			})
		})
	})

	Convey("given a 200 status with empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, []byte{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetDatasetCurrentAndNext is called", func() {
			_, err := datasetClient.GetDatasetCurrentAndNext(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/datasets/123")
			})
		})
	})

}

func TestClient_GetInstance(t *testing.T) {

	Convey("given a 200 status with valid empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, Instance{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstance is called", func() {
			instance, err := datasetClient.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, Instance{})
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123")
			})
		})
	})

	Convey("given a 200 status with empty body is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, []byte{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstance is called", func() {
			_, err := datasetClient.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123")
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := &dphttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("you aint seen me right"))),
				}, nil
			},
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{"/healthcheck"}
			},
		}

		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstance is called", func() {
			_, err := datasetClient.GetInstance(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123, body: you aint seen me right").Error())
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123")
			})
		})
	})
}

func TestClient_GetInstanceDimensionsBytes(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, ""})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstanceDimensionsBytes is called", func() {
			_, err := datasetClient.GetInstanceDimensionsBytes(ctx, userAuthToken, serviceAuthToken, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123/dimensions")
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := &dphttp.ClienterMock{
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("resource not found"))),
				}, nil
			},
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{"/healthcheck"}
			},
		}

		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstanceDimensionsBytes is called", func() {
			_, err := datasetClient.GetInstanceDimensionsBytes(ctx, userAuthToken, serviceAuthToken, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123/dimensions, body: resource not found").Error())
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123/dimensions")
			})
		})
	})
}

func TestClient_GetInstances(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, Instance{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstance is called", func() {
			_, err := datasetClient.GetInstances(ctx, userAuthToken, serviceAuthToken, collectionID, url.Values{})

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances")
			})
		})

		Convey("When GetInstance is called with filters", func() {
			_, err := datasetClient.GetInstances(ctx, userAuthToken, serviceAuthToken, collectionID, url.Values{
				"id":      []string{"123"},
				"version": []string{"999"},
			})

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected query parameters", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances?id=123&version=999")
			})
		})
	})
}

func Test_PutInstanceImportTasks(t *testing.T) {

	data := InstanceImportTasks{
		ImportObservations: &ImportObservationsTask{State: StateSubmitted.String()},
		BuildHierarchyTasks: []*BuildHierarchyTask{
			{DimensionName: "dimension1", State: StateCompleted.String()},
			{DimensionName: "dimension2", State: StateCreated.String()},
		},
		BuildSearchIndexTasks: []*BuildSearchIndexTask{
			{State: StateSubmitted.String()},
			{State: StateCompleted.String()},
		},
	}

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		datasetClient := newDatasetClient(httpClient)

		Convey("when PutInstanceImportTasks is called", func() {
			err := datasetClient.PutInstanceImportTasks(ctx, serviceAuthToken, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPut, "/instances/123/import_tasks")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_PostInstanceDimensions(t *testing.T) {

	order := 1
	optionsToPost := OptionPost{
		Name:     "testName",
		Option:   "testOption",
		Label:    "testLabel",
		CodeList: "testCodeList",
		Code:     "testCode",
		Order:    &order,
	}

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		datasetClient := newDatasetClient(httpClient)
		expectedPayload, err := json.Marshal(optionsToPost)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			err := datasetClient.PostInstanceDimensions(ctx, serviceAuthToken, "123", optionsToPost)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPost, "/instances/123/dimensions")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, "wrong!"})
		datasetClient := newDatasetClient(httpClient)
		expectedPayload, err := json.Marshal(optionsToPost)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			err := datasetClient.PostInstanceDimensions(ctx, serviceAuthToken, "123", optionsToPost)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123/dimensions, body: \"wrong!\"").Error())
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPost, "/instances/123/dimensions")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
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
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		datasetClient := newDatasetClient(httpClient)
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceState is called", func() {
			err := datasetClient.PutInstanceState(ctx, serviceAuthToken, "123", StateCompleted)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

}

func Test_UpdateInstanceWithNewInserts(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		datasetClient := newDatasetClient(httpClient)

		Convey("when UpdateInstanceWithNewInserts is called", func() {
			err := datasetClient.UpdateInstanceWithNewInserts(ctx, serviceAuthToken, "123", 999)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPut, "/instances/123/inserted_observations/999")
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
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		datasetClient := newDatasetClient(httpClient)
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			err := datasetClient.PutInstanceData(ctx, serviceAuthToken, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, "wrong!"})
		datasetClient := newDatasetClient(httpClient)
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			err := datasetClient.PutInstanceData(ctx, serviceAuthToken, "123", data)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from dataset api: http://localhost:8080/instances/123, body: \"wrong!\"").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkRequestBase(httpClient, http.MethodPut, "/instances/123")
				payload, err := ioutil.ReadAll(httpClient.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_GetInstanceDimensions(t *testing.T) {

	data := Dimensions{
		Items: []Dimension{
			{
				DimensionID: "dimension1",
				InstanceID:  "instance1",
				NodeID:      "node1",
				Label:       "label",
				Option:      "option",
			},
			{
				DimensionID: "dimension2",
			},
		},
	}

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, data})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstanceDimensions is called", func() {
			dimensions, err := datasetClient.GetInstanceDimensions(ctx, serviceAuthToken, "123")

			Convey("a positive response with expected dimensions is returned", func() {
				So(err, ShouldBeNil)
				So(dimensions, ShouldResemble, data)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123/dimensions")
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, nil})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetInstanceDimensions is called", func() {
			_, err := datasetClient.GetInstanceDimensions(ctx, serviceAuthToken, "123")

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, &ErrInvalidDatasetAPIResponse{
					actualCode: http.StatusNotFound,
					uri:        "http://localhost:8080/instances/123/dimensions",
					body:       "null",
				})
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkRequestBase(httpClient, http.MethodGet, "/instances/123/dimensions")
			})
		})
	})
}

func TestClient_PatchInstanceDimensionOption(t *testing.T) {

	testNodeID := "ABC"
	testOrder := 1

	Convey("given a 200 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, nil})
		datasetClient := newDatasetClient(httpClient)

		Convey("when PatchInstanceDimensionOption is called with valid updates for nodeID and order", func() {
			err := datasetClient.PatchInstanceDimensionOption(ctx, serviceAuthToken, "123", "456", "789", testNodeID, &testOrder)

			Convey("a positive response with expected dimensions is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected patch body", func() {
				checkRequestBase(httpClient, http.MethodPatch, "/instances/123/dimensions/456/options/789")
				sentPatches := getRequestPatchBody(httpClient, 0)
				So(sentPatches, ShouldHaveLength, 2)
				So(sentPatches[0].Op, ShouldEqual, dprequest.OpAdd.String())
				So(sentPatches[0].Path, ShouldEqual, "/node_id")
				So(sentPatches[0].Value, ShouldEqual, testNodeID)
				So(sentPatches[1].Op, ShouldEqual, dprequest.OpAdd.String())
				So(sentPatches[1].Path, ShouldEqual, "/order")
				So(sentPatches[1].Value, ShouldEqual, testOrder)
			})
		})

		Convey("when PatchInstanceDimensionOption is called with a valid update for nodeID only", func() {
			err := datasetClient.PatchInstanceDimensionOption(ctx, serviceAuthToken, "123", "456", "789", testNodeID, nil)

			Convey("a positive response with expected dimensions is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected patch body", func() {
				checkRequestBase(httpClient, http.MethodPatch, "/instances/123/dimensions/456/options/789")
				sentPatches := getRequestPatchBody(httpClient, 0)
				So(sentPatches, ShouldHaveLength, 1)
				So(sentPatches[0].Op, ShouldEqual, dprequest.OpAdd.String())
				So(sentPatches[0].Path, ShouldEqual, "/node_id")
				So(sentPatches[0].Value, ShouldEqual, testNodeID)
			})
		})

		Convey("when PatchInstanceDimensionOption is called with a valid update for order", func() {
			err := datasetClient.PatchInstanceDimensionOption(ctx, serviceAuthToken, "123", "456", "789", "", &testOrder)

			Convey("a positive response with expected dimensions is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected patch body", func() {
				checkRequestBase(httpClient, http.MethodPatch, "/instances/123/dimensions/456/options/789")
				sentPatches := getRequestPatchBody(httpClient, 0)
				So(sentPatches, ShouldHaveLength, 1)
				So(sentPatches[0].Op, ShouldEqual, dprequest.OpAdd.String())
				So(sentPatches[0].Path, ShouldEqual, "/order")
				So(sentPatches[0].Value, ShouldEqual, testOrder)
			})
		})

		Convey("when PatchInstanceDimensionOption is called without any update", func() {
			err := datasetClient.PatchInstanceDimensionOption(ctx, serviceAuthToken, "123", "456", "789", "", nil)

			Convey("a positive response with expected dimensions is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do call is skipped because nothing needed to be updated", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, nil})
		datasetClient := newDatasetClient(httpClient)

		Convey("when PatchInstanceDimensionOption is called", func() {
			err := datasetClient.PatchInstanceDimensionOption(ctx, serviceAuthToken, "123", "456", "789", testNodeID, &testOrder)

			Convey("then the expected error is returned", func() {
				So(err, ShouldResemble, &ErrInvalidDatasetAPIResponse{
					actualCode: http.StatusNotFound,
					uri:        "http://localhost:8080/instances/123/dimensions/456/options/789",
					body:       "null",
				})
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkRequestBase(httpClient, http.MethodPatch, "/instances/123/dimensions/456/options/789")
			})
		})
	})
}

func TestClient_GetOptions(t *testing.T) {

	instanceID := "testInstance"
	edition := "testEdition"
	version := "tetVersion"
	dimension := "testDimension"
	offset := 1
	limit := 10
	MaxIDs = func() int { return 5 }

	Convey("given a 200 status is returned", t, func() {
		testOptions := Options{
			Items: []Option{
				{
					DimensionID: dimension,
					Label:       "optionLabel",
					Option:      "testOption",
				},
				{
					DimensionID: dimension,
					Label:       "OptionWithSpecialChars",
					Option:      "90+",
				},
			},
			Count:      2,
			Offset:     offset,
			Limit:      limit,
			TotalCount: 3,
		}
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusOK, testOptions})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetOptions is called with valid values for limit and offset", func() {
			q := QueryParams{Offset: offset, Limit: limit, IDs: []string{}}
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &q)

			Convey("a positive response is returned, with the expected options", func() {
				So(err, ShouldBeNil)
				So(options, ShouldResemble, testOptions)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/dimensions/%s/options?offset=%d&limit=%d",
					instanceID, edition, version, dimension, offset, limit)
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})

		Convey("when GetOptions is called with negative offset", func() {
			q := QueryParams{Offset: -1, Limit: limit, IDs: []string{}}
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, Options{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("when GetOptions is called with negative limit", func() {
			q := QueryParams{Offset: offset, Limit: -1, IDs: []string{}}
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &q)

			Convey("the expected error is returned and http dphttpclient.Do is not called", func() {
				So(err.Error(), ShouldResemble, "negative offsets or limits are not allowed")
				So(options, ShouldResemble, Options{})
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})

		Convey("when GetOptions is called with a list of IDs containing an existing ID, along with offset and limit", func() {
			q := QueryParams{Offset: offset, Limit: limit, IDs: []string{"testOption", "somethingElse"}}
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &q)

			Convey("a positive response is returned, with the expected options", func() {
				So(err, ShouldBeNil)
				So(options, ShouldResemble, testOptions)
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI, providing the list of IDs and no offset or limit", func() {
				expectedURI := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/dimensions/%s/options?id=testOption,somethingElse",
					instanceID, edition, version, dimension)
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})

		Convey("when GetOptions is called with a list of IDs containing more items than the maximum allowed", func() {
			q := QueryParams{Offset: offset, Limit: limit, IDs: []string{"op1", "op2", "op3", "op4", "op5", "op6"}}
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &q)

			Convey("an error is returned, with the expected options", func() {
				So(err.Error(), ShouldResemble, "too many query parameters have been provided. Maximum allowed: 5")
				So(options, ShouldResemble, Options{})
			})

			Convey("and dphttpclient.Do is not called", func() {
				So(len(httpClient.DoCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		httpClient := createHTTPClientMock(MockedHTTPResponse{http.StatusNotFound, Options{}})
		datasetClient := newDatasetClient(httpClient)

		Convey("when GetOptions is called", func() {
			options, err := datasetClient.GetOptions(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, nil)

			Convey("the expected error response is returned, with an empty options struct", func() {
				So(err, ShouldResemble, &ErrInvalidDatasetAPIResponse{
					actualCode: 404,
					uri:        fmt.Sprintf("http://localhost:8080/datasets/%s/editions/%s/versions/%s/dimensions/%s/options", instanceID, edition, version, dimension),
					body:       "{\"items\":null,\"count\":0,\"offset\":0,\"limit\":0,\"total_count\":0}",
				})
				So(options, ShouldResemble, Options{})
			})

			Convey("and dphttpclient.Do is called 1 time with the expected URI", func() {
				expectedURI := fmt.Sprintf("/datasets/%s/editions/%s/versions/%s/dimensions/%s/options", instanceID, edition, version, dimension)
				checkRequestBase(httpClient, http.MethodGet, expectedURI)
			})
		})
	})
}

func TestClient_GetOptionsInBatches(t *testing.T) {

	instanceID := "testInstance"
	edition := "testEdition"
	version := "tetVersion"
	dimension := "testDimension"

	opts0 := Options{
		Items: []Option{
			{DimensionID: "testDimension", Label: "Option one", Option: "op1"},
			{DimensionID: "testDimension", Label: "Option two", Option: "op2"}},
		Count:      2,
		TotalCount: 3,
		Limit:      2,
		Offset:     0,
	}

	opts1 := Options{
		Items: []Option{
			{DimensionID: "testDimension", Label: "Option three", Option: "op3"}},
		Count:      1,
		TotalCount: 3,
		Limit:      2,
		Offset:     2,
	}
	batchSize := 2
	maxWorkers := 1

	Convey("When a 200 OK status is returned in 2 consecutive calls", t, func() {

		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusOK, opts0},
			MockedHTTPResponse{http.StatusOK, opts1})
		datasetClient := newDatasetClient(httpClient)

		// testProcess is a generic batch processor for testing
		processedBatches := []Options{}
		var testProcess OptionsBatchProcessor = func(batch Options) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetOptionsInBatches succeeds and returns the accumulated items from all the batches", func() {
			opts, err := datasetClient.GetOptionsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, batchSize, maxWorkers)

			So(err, ShouldBeNil)
			So(opts, ShouldResemble, Options{
				Items: []Option{
					{DimensionID: "testDimension", Label: "Option one", Option: "op1"},
					{DimensionID: "testDimension", Label: "Option two", Option: "op2"},
					{DimensionID: "testDimension", Label: "Option three", Option: "op3"}},
				Count:      3,
				TotalCount: 3,
				Limit:      0,
				Offset:     0,
			})
		})

		Convey("then GetOptionsBatchProcess calls the batchProcessor function twice, with the expected baches", func() {
			err := datasetClient.GetOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, nil, testProcess, batchSize, maxWorkers)
			So(err, ShouldBeNil)
			So(processedBatches, ShouldResemble, []Options{opts0, opts1})
			So(httpClient.DoCalls(), ShouldHaveLength, 2)
			So(httpClient.DoCalls()[0].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=0&limit=2")
			So(httpClient.DoCalls()[1].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=2&limit=2")
		})

		Convey("and a list of IDs is provided, then GetOptionsBatchProcess calls the batchProcessor function for the expected baches of IDs", func() {
			err := datasetClient.GetOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, &[]string{"op1", "op3", "op5"}, testProcess, batchSize, maxWorkers)
			So(err, ShouldBeNil)
			So(processedBatches, ShouldHaveLength, 2)
			So(httpClient.DoCalls(), ShouldHaveLength, 2)
			So(httpClient.DoCalls()[0].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?id=op1,op3")
			So(httpClient.DoCalls()[1].Req.URL.String(), ShouldResemble,
				"http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?id=op5")
		})
	})

	Convey("When a 400 error status is returned in the first call", t, func() {
		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusBadRequest, ""})
		datasetClient := newDatasetClient(httpClient)

		// testProcess is a generic batch processor for testing
		processedBatches := []Options{}
		var testProcess OptionsBatchProcessor = func(batch Options) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetOptionsInBatches fails with the expected error and the process is aborted", func() {
			_, err := datasetClient.GetOptionsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=0&limit=2")
		})

		Convey("then GetOptionsBatchProcess fails with the expected error and doesn't call the batchProcessor", func() {
			err := datasetClient.GetOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, nil, testProcess, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=0&limit=2")
			So(processedBatches, ShouldResemble, []Options{})
		})
	})

	Convey("When a 200 error status is returned in the first call and 400 error is returned in the second call", t, func() {
		httpClient := createHTTPClientMock(
			MockedHTTPResponse{http.StatusOK, opts0},
			MockedHTTPResponse{http.StatusBadRequest, ""})
		datasetClient := newDatasetClient(httpClient)

		// testProcess is a generic batch processor for testing
		processedBatches := []Options{}
		var testProcess OptionsBatchProcessor = func(batch Options) (abort bool, err error) {
			processedBatches = append(processedBatches, batch)
			return false, nil
		}

		Convey("then GetOptionsInBatches fails with the expected error, corresponding to the second batch, and the process is aborted", func() {
			_, err := datasetClient.GetOptionsInBatches(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=2&limit=2")
		})

		Convey("then GetOptionsBatchProcess fails with the expected error and calls the batchProcessor for the first batch only", func() {
			err := datasetClient.GetOptionsBatchProcess(ctx, userAuthToken, serviceAuthToken, collectionID, instanceID, edition, version, dimension, nil, testProcess, batchSize, maxWorkers)
			So(err.(*ErrInvalidDatasetAPIResponse).actualCode, ShouldEqual, http.StatusBadRequest)
			So(err.(*ErrInvalidDatasetAPIResponse).uri, ShouldResemble, "http://localhost:8080/datasets/testInstance/editions/testEdition/versions/tetVersion/dimensions/testDimension/options?offset=2&limit=2")
			So(processedBatches, ShouldResemble, []Options{opts0})
		})
	})

}

func newDatasetClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, httpClient)
	datasetClient := NewWithHealthClient(healthClient)
	return datasetClient
}

func createHTTPClientMock(mockedHTTPResponse ...MockedHTTPResponse) *dphttp.ClienterMock {
	numCall := 0
	return &dphttp.ClienterMock{
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			body, _ := json.Marshal(mockedHTTPResponse[numCall].Body)
			statusCode := mockedHTTPResponse[numCall].StatusCode
			numCall++
			return &http.Response{
				StatusCode: statusCode,
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
