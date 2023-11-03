package nlp

import (
	"net/http"
	"net/http/httptest"
)

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

// mockScrubberServer creates and returns a mock HTTP test server
// that responds with a predefined JSON structure simulating a Scrubber API response.
func mockScrubberServer() *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseJSON := `
		{
			"time": "testTime",
			"query": "testQuery",
			"results": {
			  "areas": [
				{
				  "name": "City of London",
				  "region": "London",
				  "region_code": "E12000007",
				  "codes": {
					"E00000018": "E00000018"
				  }
				}
			  ],
			  "industries": [
				{
				  "code": "01230",
				  "name": "Growing of citrus fruits"
				}
			  ]
			}
		}
		`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))

	return testServer
}

// mockCategoryServer creates and returns a mock HTTP test server
// that responds with a predefined JSON structure simulating a Category API response.
func mockCategoryServer() *httptest.Server {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseJSON := `
		[
			{
				"s": 0.6395713392217672,
				"c": [
					"peoplepopulationandcommunity",
					"healthandsocialcare",
					"conditionsanddiseases"
				]
			},
			{
				"s": 0.6393863260746002,
				"c": [
					"peoplepopulationandcommunity",
					"healthandsocialcare",
					"healthcaresystem"
				]
			}
		]
		`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))

	return testServer
}
