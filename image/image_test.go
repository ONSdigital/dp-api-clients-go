package image

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"

	"github.com/pkg/errors"
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
	So(mockdphttpCli.DoCalls()[0].Req.Header[dphttp.AuthHeaderKey][0], ShouldEqual, "Bearer "+serviceAuthToken)
	So(mockdphttpCli.DoCalls()[0].Req.Header[dphttp.FlorenceHeaderKey][0], ShouldEqual, userAuthToken)
	So(mockdphttpCli.DoCalls()[0].Req.Header[dphttp.CollectionIDHeaderKey][0], ShouldEqual, collectionID)
}

func createHTTPClientMock(retCode int, body []byte) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{}
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: retCode,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		},
	}
}

func TestClient_New(t *testing.T) {
	Convey("NewAPIClient creates a new API client with the expected URL and name", t, func() {
		imageClient := NewAPIClient(testHost)
		So(imageClient.URL(), ShouldEqual, testHost)
		So(imageClient.HealthClient().Name, ShouldEqual, "image-api")
	})

	Convey("Given an existing healthcheck client", t, func() {
		hcClient := health.NewClient("generic", testHost)
		Convey("The creating a new iamge API client providing it, results in a new client with the expected URL and name", func() {
			imageClient := NewWithHealthClient(hcClient)
			So(imageClient.URL(), ShouldEqual, testHost)
			So(imageClient.HealthClient().Name, ShouldEqual, "image-api")
		})
	})
}

func createImageAPIWithClienter(clienter dphttp.Clienter) *Client {
	hcCli := health.NewClientWithClienter("", testHost, clienter)
	return NewWithHealthClient(hcCli)
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	pathHealth := "/health"
	pathHealthcheck := "/healthcheck"

	Convey("given a clienter mock without an empty list of paths with no retry", t, func() {

		clienter := &dphttp.ClienterMock{
			SetPathsWithNoRetriesFunc: func(paths []string) {
				return
			},
			GetPathsWithNoRetriesFunc: func() []string {
				return []string{}
			},
		}

		Convey("and clienter.Do returns an error", func() {
			clientError := errors.New("disciples of the watch obey")
			clienter.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{}, clientError
			}
			clienter.SetPathsWithNoRetries([]string{pathHealth, pathHealthcheck})

			hcCli := health.NewClientWithClienter("", testHost, clienter)
			imageClient := NewWithHealthClient(hcCli)
			check := initialState

			Convey("when imageClient.Checker is called", func() {
				err := imageClient.Checker(ctx, &check)
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
					So(doCalls[0].Req.URL.Path, ShouldEqual, pathHealth)
				})
			})
		})

		Convey("and clienter.Do returns 500 response", func() {
			clienter.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
				}, nil
			}
			clienter.SetPathsWithNoRetries([]string{pathHealth, pathHealthcheck})

			imageClient := createImageAPIWithClienter(clienter)
			check := initialState

			Convey("when imageClient.Checker is called", func() {
				err := imageClient.Checker(ctx, &check)
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
					So(doCalls[0].Req.URL.Path, ShouldEqual, pathHealth)
				})
			})
		})

		Convey("and clienter.Do returns 404 response", func() {
			clienter.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
				}, nil
			}
			clienter.SetPathsWithNoRetries([]string{pathHealth, pathHealthcheck})

			imageClient := createImageAPIWithClienter(clienter)
			check := initialState

			Convey("when imageClient.Checker is called", func() {
				err := imageClient.Checker(ctx, &check)
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
					So(doCalls[0].Req.URL.Path, ShouldEqual, pathHealth)
					So(doCalls[1].Req.URL.Path, ShouldEqual, pathHealthcheck)
				})
			})
		})

		Convey("and clienter.Do returns 429 response", func() {
			clienter.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 429,
				}, nil
			}
			clienter.SetPathsWithNoRetries([]string{pathHealth, pathHealthcheck})

			imageClient := createImageAPIWithClienter(clienter)
			check := initialState

			Convey("when imageClient.Checker is called", func() {
				err := imageClient.Checker(ctx, &check)
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
					So(doCalls[0].Req.URL.Path, ShouldEqual, pathHealth)
				})
			})
		})

		Convey("and clienter.Do returns 200 response", func() {
			clienter.DoFunc = func(ctx context.Context, req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
				}, nil
			}
			clienter.SetPathsWithNoRetries([]string{pathHealth, pathHealthcheck})

			imageClient := createImageAPIWithClienter(clienter)
			check := initialState

			Convey("when imageClient.Checker is called", func() {
				err := imageClient.Checker(ctx, &check)
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
					So(doCalls[0].Req.URL.Path, ShouldEqual, pathHealth)
				})
			})
		})
	})
}

func TestClient_GetImages(t *testing.T) {
	Convey("given a 200 status is returned with an empty result list", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/images_0.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when GetImages is called", func() {
			m, err := cli.GetImages(ctx, userAuthToken, serviceAuthToken, collectionID)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(m.Count, ShouldEqual, 0)
				So(m.Items, ShouldBeEmpty)
				So(m.Limit, ShouldEqual, 0)
				So(m.Offset, ShouldEqual, 0)
				So(m.TotalCount, ShouldEqual, 0)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/images")
			})
		})
	})

	Convey("given a 200 status is returned with an single result list", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/images_1.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when GetImages is called", func() {
			m, err := cli.GetImages(ctx, userAuthToken, serviceAuthToken, collectionID)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(m.Count, ShouldEqual, 1)
				So(m.Items, ShouldNotBeEmpty)
				So(m.Items, ShouldHaveLength, 1)
				mItem := m.Items[0]
				So(mItem.Id, ShouldResemble, "042e216a-7822-4fa0-a3d6-e3f5248ffc35")
				So(mItem.Downloads, ShouldNotBeEmpty)
				So(m.Limit, ShouldEqual, 1)
				So(m.Offset, ShouldEqual, 1)
				So(m.TotalCount, ShouldEqual, 2)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/images")
			})
		})
	})
}

func TestClient_PostImage(t *testing.T) {

	newImage := NewImage{
		CollectionId: "123",
		Filename:     "pinguino.png",
		License: License{
			Title: "Some licence",
			Href:  "http://lic/lic",
		},
		Type: "animal",
	}

	Convey("given a 200 status is returned", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/image.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := createImageAPIWithClienter(mockdphttpCli)
		expectedPayload, err := json.Marshal(newImage)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			m, err := cli.PostImage(ctx, userAuthToken, serviceAuthToken, collectionID, newImage)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(m.Id, ShouldResemble, "042e216a-7822-4fa0-a3d6-e3f5248ffc35")
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images")
				payload, err := ioutil.ReadAll(mockdphttpCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)
		expectedPayload, err := json.Marshal(newImage)
		So(err, ShouldBeNil)

		Convey("when PostInstanceDimensions is called", func() {
			_, err := cli.PostImage(ctx, userAuthToken, serviceAuthToken, collectionID, newImage)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images")
				payload, err := ioutil.ReadAll(mockdphttpCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_GetImage(t *testing.T) {
	Convey("given a 200 status is returned", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/image.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when GetImages is called", func() {
			m, err := cli.GetImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(m.Id, ShouldResemble, "042e216a-7822-4fa0-a3d6-e3f5248ffc35")

			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/images/123")
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("resource not found"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when GetInstanceDimensionsBytes is called", func() {
			_, err := cli.GetImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123, body: resource not found").Error())
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodGet, "/images/123")
			})
		})
	})
}

func TestClient_PutImage(t *testing.T) {

	data := Image{
		Id:           "123",
		CollectionId: "123",
		State:        "created",
		Filename:     "pinguino.png",
		License: License{
			Title: "Some licence",
			Href:  "http://lic/lic",
		},
		Upload: ImageUpload{
			Path: "http://s3bucket/abcd.png",
		},
		Type:      "animals",
		Downloads: map[string]ImageDownload{},
	}

	Convey("given a 200 status is returned", t, func() {
		searchResp, err := ioutil.ReadFile("./response_mocks/image.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, searchResp)
		cli := createImageAPIWithClienter(mockdphttpCli)
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			m, err := cli.PutImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
				So(m.Id, ShouldResemble, "042e216a-7822-4fa0-a3d6-e3f5248ffc35")
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPut, "/images/123")
				payload, err := ioutil.ReadAll(mockdphttpCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)
		expectedPayload, err := json.Marshal(data)
		So(err, ShouldBeNil)

		Convey("when PutInstanceData is called", func() {
			_, err := cli.PutImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123", data)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPut, "/images/123")
				payload, err := ioutil.ReadAll(mockdphttpCli.DoCalls()[0].Req.Body)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, expectedPayload)
			})
		})
	})
}

func TestClient_PostImageUpload(t *testing.T) {

	data := ImageUpload{
		Path: "http://s3bucket/abcd.png",
	}

	Convey("given a 200 status is returned", t, func() {

		mockdphttpCli := createHTTPClientMock(http.StatusOK, []byte{})
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when ImportDownloadVariant is called", func() {
			err := cli.PostImageUpload(ctx, userAuthToken, serviceAuthToken, collectionID, "123", data)

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/upload")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when ImportDownloadVariant is called", func() {
			err := cli.PostImageUpload(ctx, userAuthToken, serviceAuthToken, collectionID, "123", data)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123/upload, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/upload")
			})
		})
	})
}

func TestClient_PublishImage(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {

		mockdphttpCli := createHTTPClientMock(http.StatusOK, []byte{})
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when PublishImage is called", func() {
			err := cli.PublishImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/publish")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when PublishImage is called", func() {
			err := cli.PublishImage(ctx, userAuthToken, serviceAuthToken, collectionID, "123")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123/publish, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/publish")
			})
		})
	})
}

func TestClient_PutDownloadVariant(t *testing.T) {

	w := 222
	h := 333
	data := ImageDownload{
		Size:    111,
		Type:    "downloadType",
		Width:   &w,
		Height:  &h,
		Private: "myImage",
	}

	Convey("given a 200 status is returned", t, func() {
		mockDownloadVariant, err := ioutil.ReadFile("./response_mocks/download.json")
		So(err, ShouldBeNil)

		mockdphttpCli := createHTTPClientMock(http.StatusOK, mockDownloadVariant)
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when PutDownloadVariant is called", func() {
			ret, err := cli.PutDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original", data)

			Convey("a positive response is returned with the expected updated ImageDownload", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldResemble, ImageDownload{
					Size: 1024,
					Href: "http://download.ons.gov.uk/images/042e216a-7822-4fa0-a3d6-e3f5248ffc35/image-name.png",
				})
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPut, "/images/123/downloads/original")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when ImportDownloadVariant is called", func() {
			ret, err := cli.PutDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original", data)

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123/downloads/original, body: wrong!").Error())
				So(ret, ShouldResemble, ImageDownload{})
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPut, "/images/123/downloads/original")
			})
		})
	})
}

func TestClient_ImportDownloadVariant(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {

		mockdphttpCli := createHTTPClientMock(http.StatusOK, []byte{})
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when ImportDownloadVariant is called", func() {
			err := cli.ImportDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/downloads/original/import")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when ImportDownloadVariant is called", func() {
			err := cli.ImportDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123/downloads/original/import, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/downloads/original/import")
			})
		})
	})
}

func TestClient_CompleteDownloadVariant(t *testing.T) {

	Convey("given a 200 status is returned", t, func() {

		mockdphttpCli := createHTTPClientMock(http.StatusOK, []byte{})
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when CompleteDownloadVariant is called", func() {
			err := cli.CompleteDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original")

			Convey("a positive response is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and dphttpclient.Do is called 1 time", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/downloads/original/complete")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("given a 404 status is returned", t, func() {
		mockdphttpCli := createHTTPClientMock(http.StatusNotFound, []byte("wrong!"))
		cli := createImageAPIWithClienter(mockdphttpCli)

		Convey("when CompleteDownloadVariant is called", func() {
			err := cli.CompleteDownloadVariant(ctx, userAuthToken, serviceAuthToken, collectionID, "123", "original")

			Convey("then the expected error is returned", func() {
				So(err.Error(), ShouldResemble, errors.Errorf("invalid response: 404 from image api: http://localhost:8080/images/123/downloads/original/complete, body: wrong!").Error())
			})

			Convey("and dphttpclient.Do is called 1 time with expected parameters", func() {
				checkResponseBase(mockdphttpCli, http.MethodPost, "/images/123/downloads/original/complete")
			})
		})
	})
}
