package cantabularmetadata_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabularmetadata"

	dphttp "github.com/ONSdigital/dp-net/v3/http"
)

// newMockedClient creates a new client with a mocked response for post requests,
// according to the provided response string and status code.
func newMockedClient(response string, statusCode int) (*dphttp.ClienterMock, *cantabularmetadata.Client) {
	httpClient := &dphttp.ClienterMock{
		PostFunc: func(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
			return Response(
				[]byte(response),
				statusCode,
			), nil
		},
	}

	client := cantabularmetadata.NewClient(
		cantabularmetadata.Config{
			Host: "cantabular.metadata.host",
		},
		httpClient,
	)

	return httpClient, client
}

func Response(body []byte, statusCode int) *http.Response {
	reader := bytes.NewBuffer(body)
	readCloser := ioutil.NopCloser(reader)

	return &http.Response{
		StatusCode: statusCode,
		Body:       readCloser,
	}
}

func testErrorResponse(errorMsg string) []byte {
	return []byte(fmt.Sprintf(`{"message":"%s"}`, errorMsg))
}
