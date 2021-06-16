package cantabular_test

import (
	"bytes"
	"fmt"
	"errors"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	dphttp "github.com/ONSdigital/dp-net/http"
	dperrors "github.com/ONSdigital/dp-api-clients-go/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCodebookUnhappy(t *testing.T) {

	Convey("Given a Cantabular returns an error response from the /Codebook endpoint", t, func() {
		testCtx := context.Background()

		errorMessage := "this is what cantabular errors look like"

		mockHttpClient := &dphttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return Response(testErrorResponse(errorMessage), http.StatusNotFound), nil
			},
		}

		cantabularClient := cantabular.NewClient(
			mockHttpClient,
			cantabular.Config{},
		)

		Convey("When the GetCodebook method is called", func() {
			req := cantabular.GetCodebookRequest{}
			cb, err := cantabularClient.GetCodebook(testCtx, req)

			Convey("Then returned error messaage should be extracted correctly", func() {
				So(cb, ShouldBeNil)
				So(err.Error(), ShouldEqual, errorMessage)
			})
		})
	})

	Convey("Given a Cantabular fails to respond", t, func() {
		testCtx := context.Background()

		mockHttpClient := &dphttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return nil, errors.New("server failed to respond")
			},
		}

		cantabularClient := cantabular.NewClient(
			mockHttpClient,
			cantabular.Config{},
		)

		Convey("When the GetCodebook method is called", func() {
			req := cantabular.GetCodebookRequest{}
			cb, err := cantabularClient.GetCodebook(testCtx, req)

			Convey("Then the status code 500 should be recoverable from the error", func() {
				So(cb, ShouldBeNil)
				So(dperrors.StatusCode(err), ShouldEqual, http.StatusInternalServerError)
			})
		})
	})
}

func TestGetCodebookHappy(t *testing.T) {

	Convey("Given a correct response from the /Codebook endpoint", t, func() {
		testCtx := context.Background()

		resp, err := testCodebookResponse()
		So(err, ShouldBeNil)

		mockHttpClient := &dphttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return Response(
					resp,
					http.StatusOK,
				), nil
			},
		}

		cantabularClient := cantabular.NewClient(
			mockHttpClient,
			cantabular.Config{},
		)

		Convey("When the GetCodebook method is called", func() {
			req := cantabular.GetCodebookRequest{}
			cb, err := cantabularClient.GetCodebook(testCtx, req)
			So(err, ShouldBeNil)

			Convey("Then the expected codebook information should be returned", func() {
				So(cb.Dataset.Name,                          ShouldEqual,      "Example")
				So(cb.Codebook,                              ShouldHaveLength,  5)
				So(cb.Codebook[0].Name,                      ShouldEqual,      "city")
				So(cb.Codebook[1].Labels[0],                 ShouldEqual,      "England")
				So(cb.Codebook[2].Labels,                    ShouldHaveLength,  2)
				So(cb.Codebook[3].Codes[2],                  ShouldEqual,      "2")
				So(cb.Codebook[4].MapFrom[0].SourceNames[0], ShouldEqual,      "siblings")
				So(cb.Codebook[4].MapFrom[0].Code[1],        ShouldEqual,      "1-2")
			})
		})
	})
}

func Response(body []byte, statusCode int) *http.Response {
	reader := bytes.NewBuffer(body)
	readCloser := ioutil.NopCloser(reader)

	return &http.Response{
		StatusCode: statusCode,
		Body:       readCloser,
	}
}

func testErrorResponse(errorMsg string) []byte{
	return []byte(fmt.Sprintf(`{"message":"%s"}`, errorMsg))
}

func testCodebookResponse() ([]byte, error){
	b, err := ioutil.ReadFile("codebook_test.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}

	return b, nil
}