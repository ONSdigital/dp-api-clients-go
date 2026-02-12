package zebedee

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-mocking/httpmocks"
	dprequest "github.com/ONSdigital/dp-net/v3/request"
	. "github.com/smartystreets/goconvey/convey"
)

func TestClientGetCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expectedPath := "/collectionDetails/" + testCollectionID
	expectedCollection := Collection{
		ID:             "collection-id",
		Name:           "Example collection",
		ApprovalStatus: "APPROVED",
		Type:           "collection",
		Inprogress: []CollectionItem{{
			ID:    "item-1",
			State: "InProgress",
			Title: "Item title",
			URI:   "/item/1",
		}},
	}
	responseBody, _ := json.Marshal(expectedCollection)

	Convey("given a 200 response", t, func() {
		body := httpmocks.NewReadCloserMock(responseBody, nil)
		response := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when GetCollection is called", func() {
			collection, err := zebedeeClient.GetCollection(ctx, testAccessToken, testCollectionID)

			Convey("then the expected collection is returned", func() {
				So(err, ShouldBeNil)
				So(collection, ShouldResemble, expectedCollection)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.Method, ShouldEqual, http.MethodGet)
				So(doCalls[0].Req.URL.Path, ShouldEqual, expectedPath)
				So(doCalls[0].Req.Header.Get(dprequest.FlorenceHeaderKey), ShouldEqual, testAccessToken)
			})
		})
	})

	Convey("given a 500 response", t, func() {
		body := httpmocks.NewReadCloserMock([]byte(`{}`), nil)
		response := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when GetCollection is called", func() {
			collection, err := zebedeeClient.GetCollection(ctx, testAccessToken, testCollectionID)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
				So(collection, ShouldResemble, Collection{})
			})
		})
	})

	Convey("given a 200 response with invalid JSON", t, func() {
		body := httpmocks.NewReadCloserMock([]byte(`invalid-json`), nil)
		response := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when GetCollection is called", func() {
			collection, err := zebedeeClient.GetCollection(ctx, testAccessToken, testCollectionID)

			Convey("then a JSON error is returned", func() {
				So(err, ShouldNotBeNil)
				So(collection, ShouldResemble, Collection{})
			})
		})
	})
}

func TestClientCreateCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expectedPath := "/collection"
	collection := Collection{
		Name: "Create collection",
		Type: "collection",
		Inprogress: []CollectionItem{{
			ID:    "item-1",
			State: "InProgress",
			Title: "Item title",
			URI:   "/item/1",
		}},
	}
	createdCollection := Collection{
		ID:             "collection-id",
		Name:           "Create collection",
		ApprovalStatus: "APPROVED",
		Type:           "collection",
	}
	responseBody, _ := json.Marshal(createdCollection)

	Convey("given a 201 response", t, func() {
		body := httpmocks.NewReadCloserMock(responseBody, nil)
		response := httpmocks.NewResponseMock(body, http.StatusCreated)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when CreateCollection is called", func() {
			created, err := zebedeeClient.CreateCollection(ctx, testAccessToken, collection)

			Convey("then the expected collection is returned", func() {
				So(err, ShouldBeNil)
				So(created, ShouldResemble, createdCollection)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.Method, ShouldEqual, http.MethodPost)
				So(doCalls[0].Req.URL.Path, ShouldEqual, expectedPath)
				So(doCalls[0].Req.Header.Get(dprequest.FlorenceHeaderKey), ShouldEqual, testAccessToken)

				var payload Collection
				err := json.NewDecoder(doCalls[0].Req.Body).Decode(&payload)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, collection)
			})
		})
	})

	Convey("given a 500 response", t, func() {
		body := httpmocks.NewReadCloserMock([]byte(`{}`), nil)
		response := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when CreateCollection is called", func() {
			created, err := zebedeeClient.CreateCollection(ctx, testAccessToken, collection)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
				So(created, ShouldResemble, Collection{})
			})
		})
	})

	Convey("given a 200 response with invalid JSON", t, func() {
		body := httpmocks.NewReadCloserMock([]byte(`invalid-json`), nil)
		response := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when CreateCollection is called", func() {
			created, err := zebedeeClient.CreateCollection(ctx, testAccessToken, collection)

			Convey("then a JSON error is returned", func() {
				So(err, ShouldNotBeNil)
				So(created, ShouldResemble, Collection{})
			})
		})
	})
}

func TestClientSaveContentToCollection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pagePath := "some/page"
	expectedPath := "/content/" + testCollectionID
	expectedQuery := "uri=" + pagePath + "/data.json"
	content := map[string]interface{}{
		"type":  "bulletin",
		"title": "Title",
	}
	responseBody := []byte(`{}`)

	Convey("given a 201 response", t, func() {
		body := httpmocks.NewReadCloserMock(responseBody, nil)
		response := httpmocks.NewResponseMock(body, http.StatusCreated)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when SaveContentToCollection is called", func() {
			err := zebedeeClient.SaveContentToCollection(ctx, testAccessToken, testCollectionID, pagePath, content)

			Convey("then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.Method, ShouldEqual, http.MethodPost)
				So(doCalls[0].Req.URL.Path, ShouldEqual, expectedPath)
				So(doCalls[0].Req.URL.RawQuery, ShouldEqual, expectedQuery)
				So(doCalls[0].Req.Header.Get(dprequest.FlorenceHeaderKey), ShouldEqual, testAccessToken)

				var payload map[string]interface{}
				err := json.NewDecoder(doCalls[0].Req.Body).Decode(&payload)
				So(err, ShouldBeNil)
				So(payload, ShouldResemble, content)
			})
		})
	})

	Convey("given content that cannot be marshalled", t, func() {
		httpClient := newMockHTTPClient(&http.Response{}, nil)
		zebedeeClient := newZebedeeClient(httpClient)
		invalidContent := map[string]interface{}{
			"bad": make(chan int),
		}

		Convey("when SaveContentToCollection is called", func() {
			err := zebedeeClient.SaveContentToCollection(ctx, testAccessToken, testCollectionID, pagePath, invalidContent)

			Convey("then a marshal error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "error while attempting to marshall content")
			})
		})
	})

	Convey("given a 500 response", t, func() {
		body := httpmocks.NewReadCloserMock(responseBody, nil)
		response := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when SaveContentToCollection is called", func() {
			err := zebedeeClient.SaveContentToCollection(ctx, testAccessToken, testCollectionID, pagePath, content)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
			})
		})
	})
}
