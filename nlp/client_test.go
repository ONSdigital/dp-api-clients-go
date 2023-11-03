package nlp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	dphttp "github.com/ONSdigital/dp-net/v2/http"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBuildURL(t *testing.T) {
	baseURL := "https://example.com/api"
	queryKey := "search"
	Convey("Given a base URL, query parameters, and a query key", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, "example", queryKey)

			Convey("The URL should be built correctly", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=example")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an invalid base URL", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(":", "example", queryKey)

			Convey("An error should be returned", func() {
				So(resultURL, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an empty query parameter", t, func() {
		Convey("When buildURL is called", func() {
			resultURL, err := buildURL(baseURL, "", queryKey)

			Convey("The URL should only contain the query key", func() {
				So(resultURL.String(), ShouldEqual, "https://example.com/api?search=")
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetBerlin(t *testing.T) {
	// Initialize a testServer with a predefined JSON response
	testServer := mockBerlinServer()

	defer testServer.Close() // Close the test server when done

	// Initialize a Client instance
	client := &Client{
		berlinBaseURL:  testServer.URL,
		berlinEndpoint: "/v1/berlin",
		client:         dphttp.Client{}, // Ensure you set this up appropriately
	}

	Convey("Test GetBerlin method", t, func() {
		Convey("When making a successful request to get Berlin data", func() {
			ctx := context.Background()
			query := "exampleQuery"

			Convey("It should return Berlin data without error", func() {
				berlin, err := client.GetBerlin(ctx, query)

				So(err, ShouldBeNil)
				So(berlin, ShouldNotBeNil)
				So(berlin.Matches[0].Codes[0], ShouldEqual, "codeTest")
				So(berlin.Matches[0].Codes[1], ShouldEqual, "codeTest")
				So(berlin.Matches[0].Encoding, ShouldEqual, "encodingTest")
				So(berlin.Matches[0].ID, ShouldEqual, "idTest")
				So(berlin.Matches[0].Key, ShouldEqual, "keyTest")
				So(berlin.Matches[0].Names[0], ShouldEqual, "nameTest")
				So(berlin.Matches[0].Names[1], ShouldEqual, "nameTest")
				So(berlin.Matches[0].State[0], ShouldEqual, "stateTest")
				So(berlin.Matches[0].Subdivision[0], ShouldEqual, "subDivTest")
				So(berlin.Matches[0].Words[0], ShouldEqual, "wordsTest")
			})
		})

		Convey("When encountering an error during the request", func() {
			// Mock the request to intentionally cause an error
			client.berlinBaseURL = "invalidURL" // Simulate an invalid URL

			ctx := context.Background()
			query := "exampleQuery"

			Convey("It should return an error", func() {
				berlin, err := client.GetBerlin(ctx, query)

				t.Logf("With a view %s", err.Error())
				So(err, ShouldNotBeNil)
				So(berlin.Matches, ShouldBeEmpty)
			})
		})

	})
}

// mockBerlinServer creates and returns a mock HTTP test server
// that responds with a predefined JSON structure simulating a Berlin API response.
func mockBerlinServer() *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseJSON := `{
			"matches": [
				{
					"codes": [
						"codeTest",
						"codeTest"
					],
					"encoding": "encodingTest",
					"id": "idTest",
					"key": "keyTest",
					"names": [
						"nameTest",
						"nameTest"
					],
					"state": [
						"stateTest"
					],
					"subdiv": [
						"subDivTest"
					],
					"words": [
						"wordsTest"
					]
				}
			]
		  }`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))

	return testServer
}
