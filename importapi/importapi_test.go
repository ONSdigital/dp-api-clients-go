package importapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	dphttp "github.com/ONSdigital/dp-net/http"
	. "github.com/smartystreets/goconvey/convey"
)

const serviceToken = "I r a service token"

var (
	ctx          = context.Background()
	testHost     = "http://localhost:8080"
	initialState = health.CreateCheckState(service)
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")

		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		importClient := New(testHost)
		importClient.cli = clienter
		check := initialState

		Convey("when importClient.Checker is called", func() {
			err := importClient.Checker(ctx, &check)
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
			SetPathsWithNoRetriesFunc: func(paths []string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		importClient := New(testHost)
		importClient.cli = clienter
		check := initialState

		Convey("when importClient.Checker is called", func() {
			err := importClient.Checker(ctx, &check)
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
			SetPathsWithNoRetriesFunc: func(paths []string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		importClient := New(testHost)
		importClient.cli = clienter
		check := initialState

		Convey("when importClient.Checker is called", func() {
			err := importClient.Checker(ctx, &check)
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
			SetPathsWithNoRetriesFunc: func(paths []string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 429,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		importClient := New(testHost)
		importClient.cli = clienter
		check := initialState

		Convey("when importClient.Checker is called", func() {
			err := importClient.Checker(ctx, &check)
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
			SetPathsWithNoRetriesFunc: func(paths []string) {},
			DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			},
		}
		clienter.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		importClient := New(testHost)
		importClient.cli = clienter
		check := initialState

		Convey("when importClient.Checker is called", func() {
			err := importClient.Checker(ctx, &check)
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

// getMockImportAPI gets a mocked client that validates the expected request and returns the provided mocked http response.
// The request Method is always validated, and the body is validated only if it not nil
func getMockImportAPI(c C, expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectRequest.Body != nil {
			body, err := ioutil.ReadAll(r.Body)
			c.So(err, ShouldBeNil)
			expectedBody, err := ioutil.ReadAll(expectRequest.Body)
			c.So(err, ShouldBeNil)
			c.So(string(body), ShouldResemble, string(expectedBody))
		}
		c.So(r.Method, ShouldEqual, expectRequest.Method)
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))

	client := New(ts.URL)
	// Make client return on first request, no need to retry for tests
	client.cli.SetMaxRetries(0)

	return client
}

func TestGetImportJob(t *testing.T) {
	jobID := "jid1"
	jobJSON := `{"id":"` + jobID + `","links":{"instances":[{"id":"iid1","href":"iid1link"}]}}`
	jobMultiInstJSON := `{"id":"` + jobID + `","links":{"instances":[{"id":"iid1","href":"iid1link"},{"id":"iid2","href":"iid2link"}]}}`

	Convey("When no import-job is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 404, Body: ""})
		job, err := mockedAPI.GetImportJob(ctx, jobID, serviceToken)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusNotFound,
			uri:        fmt.Sprintf("%s/jobs/jid1", mockedAPI.url),
			body:       "",
		})
		So(job, ShouldResemble, ImportJob{})
	})

	Convey("When bad json is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: "oops"})
		_, err := mockedAPI.GetImportJob(ctx, jobID, serviceToken)
		So(err, ShouldNotBeNil)
		So(err, ShouldHaveSameTypeAs, &json.SyntaxError{})
	})

	Convey("When server error is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 500, Body: "[]"})
		_, err := mockedAPI.GetImportJob(ctx, jobID, serviceToken)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusInternalServerError,
			uri:        fmt.Sprintf("%s/jobs/jid1", mockedAPI.url),
			body:       "",
		})
	})

	Convey("When a single-instance import-job is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: jobJSON})
		job, err := mockedAPI.GetImportJob(ctx, jobID, serviceToken)
		So(err, ShouldBeNil)
		So(job, ShouldResemble, ImportJob{JobID: jobID, Links: LinkMap{Instances: []InstanceLink{{ID: "iid1", Link: "iid1link"}}}})
	})

	Convey("When a multiple-instance import-job is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: jobMultiInstJSON})
		job, err := mockedAPI.GetImportJob(ctx, jobID, serviceToken)
		So(err, ShouldBeNil)
		So(job, ShouldResemble, ImportJob{
			JobID: jobID,
			Links: LinkMap{
				Instances: []InstanceLink{
					{ID: "iid1", Link: "iid1link"},
					{ID: "iid2", Link: "iid2link"},
				},
			},
		})
	})
}

func TestUpdateImportJobState(t *testing.T) {
	jobID := "jid0"

	Convey("When bad request is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		err := mockedAPI.UpdateImportJobState(ctx, jobID, serviceToken, StateSubmitted)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusBadRequest,
			uri:        fmt.Sprintf("%s/jobs/jid0", mockedAPI.url),
			body:       "",
		})
	})

	Convey("When server error is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: "PUT"}, MockedHTTPResponse{StatusCode: 500, Body: "dnm"})
		err := mockedAPI.UpdateImportJobState(ctx, jobID, serviceToken, StateSubmitted)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusInternalServerError,
			uri:        fmt.Sprintf("%s/jobs/jid0", mockedAPI.url),
			body:       "",
		})
	})

	Convey("When a single import-instance is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c,
			http.Request{
				Method: "PUT",
				Body:   httpmocks.NewReadCloserMock([]byte(`{"state":"submitted"}`), nil),
			},
			MockedHTTPResponse{
				StatusCode: 200,
				Body:       "",
			},
		)
		err := mockedAPI.UpdateImportJobState(ctx, jobID, serviceToken, StateSubmitted)
		So(err, ShouldBeNil)
	})
}

func TestIncreaseProcessedInstanceCount(t *testing.T) {

	jobID := "job0"
	instanceID := "inst0"

	Convey("When bad request is returned then the expected error is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: http.MethodPut}, MockedHTTPResponse{StatusCode: http.StatusBadRequest, Body: ""})
		procInst, err := mockedAPI.IncreaseProcessedInstanceCount(ctx, jobID, serviceToken, instanceID)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusBadRequest,
			uri:        fmt.Sprintf("%s/jobs/job0/processed/inst0", mockedAPI.url),
			body:       "",
		})
		So(procInst, ShouldBeNil)
	})

	Convey("When server error is returned then the expected error is returned", t, func(c C) {
		mockedAPI := getMockImportAPI(c, http.Request{Method: http.MethodPut}, MockedHTTPResponse{StatusCode: http.StatusInternalServerError, Body: ""})
		procInst, err := mockedAPI.IncreaseProcessedInstanceCount(ctx, jobID, serviceToken, instanceID)
		So(err, ShouldNotBeNil)
		So(err, ShouldResemble, &ErrInvalidAPIResponse{
			actualCode: http.StatusInternalServerError,
			uri:        fmt.Sprintf("%s/jobs/job0/processed/inst0", mockedAPI.url),
			body:       "",
		})
		So(procInst, ShouldBeNil)
	})

	Convey("When ok response is returned the expected response is returned without error", t, func(c C) {
		bodyStr := `[
			{
				"id": "inst0",
				"required_count": 10,
				"processed_count": 2
			}
		]`
		mockedAPI := getMockImportAPI(c, http.Request{Method: http.MethodPut}, MockedHTTPResponse{StatusCode: http.StatusOK, Body: bodyStr})
		procInst, err := mockedAPI.IncreaseProcessedInstanceCount(ctx, jobID, serviceToken, instanceID)
		So(err, ShouldBeNil)
		So(procInst, ShouldResemble, []ProcessedInstances{
			{
				ID:             instanceID,
				RequiredCount:  10,
				ProcessedCount: 2,
			},
		})
	})
}

func TestState(t *testing.T) {
	Convey("State strings return the expected values", t, func() {
		s := StateCreated
		So(s.String(), ShouldResemble, "created")

		s = StateSubmitted
		So(s.String(), ShouldResemble, "submitted")

		s = StateCompleted
		So(s.String(), ShouldResemble, "completed")

		s = StateFailed
		So(s.String(), ShouldResemble, "failed")
	})

	Convey("State json return the expected json structures", t, func() {
		s := StateCreated
		So(*s.Json(), ShouldResemble, stateData{"created"})

		s = StateSubmitted
		So(*s.Json(), ShouldResemble, stateData{"submitted"})

		s = StateCompleted
		So(*s.Json(), ShouldResemble, stateData{"completed"})

		s = StateFailed
		So(*s.Json(), ShouldResemble, stateData{"failed"})
	})
}
