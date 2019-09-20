package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testHeader1 = "1234567890"
	testHeader2 = "0987654321"
)

func TestSetCollectionID(t *testing.T) {
	Convey("should return error if request is nil", t, func() {
		err := SetCollectionID(nil, "")

		So(err, ShouldResemble, errRequestNil)
	})

	Convey("should not add header if value is empty", t, func() {
		req := requestWithHeader(collectionID, "")

		err := SetCollectionID(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(collectionID), ShouldBeEmpty)
	})

	Convey("should overwrite an existing header", t, func() {
		req := requestWithHeader(collectionID, testHeader1)

		err := SetCollectionID(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(collectionID), ShouldEqual, testHeader2)
	})

	Convey("should set header if it does not already exist", t, func() {
		req := requestWithHeader(collectionID, "")

		err := SetCollectionID(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(collectionID), ShouldEqual, testHeader1)
	})
}

func TestGetCollectionID(t *testing.T) {
	Convey("should return expected error if request is nil", t, func() {
		actual, err := GetCollectionID(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return ErrHeaderNotFound if the collection ID request header is not found", t, func() {
		req := requestWithHeader(collectionID, "")

		actual, err := GetCollectionID(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return header value if present", t, func() {
		req := requestWithHeader(collectionID, testHeader1)

		actual, err := GetCollectionID(req)

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, testHeader1)
	})
}

func TestSetUserAuthToken(t *testing.T) {
	Convey("SetUserAuthToken should return error if request is nil", t, func() {
		err := SetUserAuthToken(nil, "")

		So(err, ShouldResemble, errRequestNil)
	})

	Convey("SetUserAuthToken should not add header if value is empty", t, func() {
		req := requestWithHeader(userAuthToken, "")

		err := SetUserAuthToken(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(userAuthToken), ShouldBeEmpty)
	})

	Convey("SetUserAuthToken should overwrite an existing header", t, func() {
		req := requestWithHeader(userAuthToken, testHeader1)

		err := SetUserAuthToken(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(userAuthToken), ShouldEqual, testHeader2)
	})

	Convey("SetUserAuthToken should set header if it does not already exist", t, func() {
		req := requestWithHeader(userAuthToken, "")

		err := SetUserAuthToken(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(userAuthToken), ShouldEqual, testHeader1)
	})
}

func TestGetUserAuthToken(t *testing.T) {
	Convey("GetUserAuthToken should return expected error if request is nil", t, func() {
		actual, err := GetUserAuthToken(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetUserAuthToken should return ErrHeaderNotFound if the userAuthToken request header is not found", t, func() {
		req := requestWithHeader(userAuthToken, "")

		actual, err := GetUserAuthToken(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetUserAuthToken should return header value if present", t, func() {
		req := requestWithHeader(userAuthToken, testHeader1)

		actual, err := GetUserAuthToken(req)

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, testHeader1)
	})
}

func TestSetServiceAuthToken(t *testing.T) {
	Convey("SetServiceAuthToken should return error if request is nil", t, func() {
		err := SetServiceAuthToken(nil, "")

		So(err, ShouldResemble, errRequestNil)
	})

	Convey("SetServiceAuthToken should not add header if value is empty", t, func() {
		req := requestWithHeader(serviceAuthToken, "")

		err := SetServiceAuthToken(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(serviceAuthToken), ShouldBeEmpty)
	})

	Convey("SetServiceAuthToken should overwrite an existing header", t, func() {
		req := requestWithHeader(serviceAuthToken, testHeader1)

		err := SetServiceAuthToken(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(serviceAuthToken), ShouldEqual, bearerPrefix + testHeader2)
	})

	Convey("SetServiceAuthToken should set header if it does not already exist", t, func() {
		req := requestWithHeader(serviceAuthToken, "")

		err := SetServiceAuthToken(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(serviceAuthToken), ShouldEqual, bearerPrefix + testHeader1)
	})
}

func TestGetServiceAuthToken(t *testing.T) {
	Convey("GetServiceAuthToken should return expected error if request is nil", t, func() {
		actual, err := GetServiceAuthToken(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetServiceAuthToken should return ErrHeaderNotFound if the serviceAuthToken request header is not found", t, func() {
		req := requestWithHeader(serviceAuthToken, "")

		actual, err := GetServiceAuthToken(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetServiceAuthToken should return header value if present", t, func() {
		req := requestWithHeader(serviceAuthToken, bearerPrefix + testHeader1)

		actual, err := GetServiceAuthToken(req)

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, testHeader1)
	})

	Convey("GetServiceAuthToken should return header value if it does not have the bearer prefix", t, func() {
		req := requestWithHeader(serviceAuthToken, bearerPrefix + testHeader1)

		actual, err := GetServiceAuthToken(req)

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, testHeader1)
	})
}

func requestWithHeader(key, val string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "http://localhost:456789/schwifty", nil)
	if len(val) > 0 {
		r.Header.Set(key, val)
	}
	return r
}
