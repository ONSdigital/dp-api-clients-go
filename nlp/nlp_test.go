package nlp

import (
	"context"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/mocks"
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
	testServer := mocks.MockBerlinServer()

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
				_, err := client.GetBerlin(ctx, query)

				t.Logf("With a view %s", err.Error())
				So(err, ShouldNotBeNil)
			})
		})

	})
}

func TestGetScrubber(t *testing.T) {
	// Initialize a testServer with a predefined JSON response
	testServer := mocks.MockScrubberServer()

	defer testServer.Close() // Close the test server when done

	// Initialize a Client instance
	client := &Client{
		scrubberBaseURL:  testServer.URL,
		scrubberEndpoint: "/v1/scrubber",
		client:           dphttp.Client{}, // Ensure you set this up appropriately
	}

	Convey("Test GetScrubber method", t, func() {
		Convey("When making a successful request to get scrubber data", func() {
			ctx := context.Background()
			query := "testQuery"

			Convey("It should return scrubber data without error", func() {
				scrubber, err := client.GetScrubber(ctx, query)

				So(err, ShouldBeNil)
				So(scrubber, ShouldNotBeNil)
				So(scrubber.Query, ShouldEqual, "testQuery")
				So(scrubber.Results, ShouldNotBeEmpty)
				So(scrubber.Results.Areas, ShouldNotBeEmpty)
				So(scrubber.Results.Areas[0].Name, ShouldEqual, "City of London")
				So(scrubber.Results.Areas[0].Region, ShouldEqual, "London")
				So(scrubber.Results.Areas[0].RegionCode, ShouldEqual, "E12000007")
				So(scrubber.Results.Industries, ShouldNotBeEmpty)
				So(scrubber.Results.Industries[0].Code, ShouldEqual, "01230")
				So(scrubber.Results.Industries[0].Name, ShouldEqual, "Growing of citrus fruits")
			})
		})

		Convey("When encountering an error during the request", func() {
			// Mock the request to intentionally cause an error
			client.scrubberBaseURL = "invalidURL" // Simulate an invalid URL

			ctx := context.Background()
			query := "exampleQuery"

			Convey("It should return an error", func() {
				_, err := client.GetScrubber(ctx, query)

				t.Logf("With a view %s", err.Error())
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetCategories(t *testing.T) {
	// Initialize a testServer with a predefined JSON response
	testServer := mocks.MockCategoryServer()

	defer testServer.Close() // Close the test server when done

	// Initialize a Client instance
	client := &Client{
		categoryBaseURL:  testServer.URL,
		categoryEndpoint: "/v1/categories",
		client:           dphttp.Client{}, // Ensure you set this up appropriately
	}

	Convey("Test GetCategory method", t, func() {
		Convey("When making a successful request to get scrubber data", func() {
			ctx := context.Background()
			query := "testQuery"

			Convey("It should return scrubber data without error", func() {
				cat, err := client.GetCategory(ctx, query)
				categories := *cat

				So(err, ShouldBeNil)
				So(categories, ShouldNotBeEmpty)
				So(categories[0].Code, ShouldNotBeEmpty)
				So(categories[0].Code[0], ShouldEqual, "peoplepopulationandcommunity")
				So(categories[0].Code[1], ShouldEqual, "healthandsocialcare")
				So(categories[0].Code[2], ShouldEqual, "conditionsanddiseases")
				So(categories[0].Score, ShouldEqual, 0.6395713392217672)
				So(categories[1].Code, ShouldNotBeEmpty)
				So(categories[1].Code[0], ShouldEqual, "peoplepopulationandcommunity")
				So(categories[1].Code[1], ShouldEqual, "healthandsocialcare")
				So(categories[1].Code[2], ShouldEqual, "healthcaresystem")
				So(categories[1].Score, ShouldEqual, 0.6393863260746002)
			})
		})

		Convey("When encountering an error during the request", func() {
			// Mock the request to intentionally cause an error
			client.categoryBaseURL = "invalidURL"

			ctx := context.Background()
			query := "exampleQuery"

			Convey("It should return an error", func() {
				_, err := client.GetCategory(ctx, query)

				t.Logf("With a view %s", err.Error())
				So(err, ShouldNotBeNil)
			})
		})
	})
}
