package recipe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	dperrors "github.com/ONSdigital/dp-api-clients-go/errors"
	"github.com/ONSdigital/dp-api-clients-go/health"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testServiceToken  = "bar"
	testUserAuthToken = "grault"
	testHost          = "http://localhost:8080"
)

var (
	errTest = errors.New("recipe API error")
	ctx     = context.Background()
)

// checkRequest validates request method, uri and headers
func checkRequest(httpClient *dphttp.ClienterMock, callIndex int, expectedMethod, expectedURI string) {
	So(httpClient.DoCalls()[callIndex].Req.URL.String(), ShouldEqual, expectedURI)
	So(httpClient.DoCalls()[callIndex].Req.Method, ShouldEqual, expectedMethod)
	So(httpClient.DoCalls()[callIndex].Req.Header.Get(dprequest.AuthHeaderKey), ShouldEqual, "Bearer "+testServiceToken)
}

func TestGetRecipe(t *testing.T) {
	recipeID := "testRecipe"
	recipeBody := `{"id":"` + recipeID + `", "format": "cantabular_table", "cantabular_blob": "123"}`
	expectedRecipe := Recipe{
		ID:             recipeID,
		Format:         "cantabular_table",
		CantabularBlob: "123",
	}

	Convey("Given that 200 OK is returned by recipe API with a valid recipe body", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(recipeBody))),
		}, nil)
		recipeClient := newRecipeClient(httpClient)

		Convey("Then whe GetRecipe is called, one GET /recipes/ID call is performed and the expected recipe is returned without error", func() {
			recipe, err := recipeClient.GetRecipe(ctx, testUserAuthToken, testServiceToken, recipeID)
			So(err, ShouldBeNil)
			So(*recipe, ShouldResemble, expectedRecipe)
			So(httpClient.DoCalls(), ShouldHaveLength, 1)
			checkRequest(httpClient, 0, http.MethodGet, fmt.Sprintf("%s/recipes/%s", testHost, recipeID))
		})
	})

	Convey("Given that 400 BadRequest is returned by recipe API", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
		}, nil)
		recipeClient := newRecipeClient(httpClient)

		Convey("Then whe GetRecipe is called, one GET /recipes/ID call is performed and the expected error is returned", func() {
			recipe, err := recipeClient.GetRecipe(ctx, testUserAuthToken, testServiceToken, recipeID)
			So(err, ShouldResemble, dperrors.New(
				errors.New(""),
				http.StatusBadRequest,
				nil))
			So(recipe, ShouldBeNil)
			So(httpClient.DoCalls(), ShouldHaveLength, 1)
			checkRequest(httpClient, 0, http.MethodGet, fmt.Sprintf("%s/recipes/%s", testHost, recipeID))
		})
	})

	Convey("Given an http client that fails to perform a request", t, func() {
		httpClient := newMockHTTPClient(nil, errTest)
		recipeClient := newRecipeClient(httpClient)

		Convey("Then whe GetRecipe is called, one GET /recipes/ID call is performed and the expected error is returned", func() {
			recipe, err := recipeClient.GetRecipe(ctx, testUserAuthToken, testServiceToken, recipeID)
			So(err, ShouldResemble, dperrors.New(
				errors.New("failed to get response from Recipe API: recipe API error"),
				http.StatusInternalServerError,
				nil))
			So(recipe, ShouldBeNil)
			So(httpClient.DoCalls(), ShouldHaveLength, 1)
			checkRequest(httpClient, 0, http.MethodGet, fmt.Sprintf("%s/recipes/%s", testHost, recipeID))
		})
	})

	Convey("Given that 200 OK is returned by recipe API with an invalid body", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("invalidRecipeBody"))),
		}, nil)
		recipeClient := newRecipeClient(httpClient)

		Convey("Then whe GetRecipe is called, one GET /recipes/ID call is performed and the expected error is returned", func() {
			recipe, err := recipeClient.GetRecipe(ctx, testUserAuthToken, testServiceToken, recipeID)
			So(err, ShouldResemble, dperrors.New(
				errors.New("failed to unmarshal response body: invalid character 'i' looking for beginning of value"),
				http.StatusInternalServerError,
				log.Data{"response_body": "invalidRecipeBody"}))
			So(recipe, ShouldBeNil)
			So(httpClient.DoCalls(), ShouldHaveLength, 1)
			checkRequest(httpClient, 0, http.MethodGet, fmt.Sprintf("%s/recipes/%s", testHost, recipeID))
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
			return []string{"/healthcheck"}
		},
	}
}

func newRecipeClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	return NewWithHealthClient(healthClient)
}
