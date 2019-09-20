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
		req := requestWithHeader(CollectionIDHeader, "")

		err := SetCollectionID(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDHeader), ShouldBeEmpty)
	})

	Convey("should overwrite an existing header", t, func() {
		req := requestWithHeader(CollectionIDHeader, testHeader1)

		err := SetCollectionID(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDHeader), ShouldEqual, testHeader2)
	})

	Convey("should set header if it does not already exist", t, func() {
		req := requestWithHeader(CollectionIDHeader, "")

		err := SetCollectionID(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDHeader), ShouldEqual, testHeader1)
	})
}

func TestGetCollectionID(t *testing.T) {
	Convey("should return expected error if request is nil", t, func() {
		actual, err := GetCollectionID(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return ErrHeaderNotFound if the collection ID request header is not found", t, func() {
		req := requestWithHeader(CollectionIDHeader, "")

		actual, err := GetCollectionID(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return header value if present", t, func() {
		req := requestWithHeader(CollectionIDHeader, testHeader1)

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
		req := requestWithHeader(UserAuthTokenHeader, "")

		err := SetUserAuthToken(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(UserAuthTokenHeader), ShouldBeEmpty)
	})

	Convey("SetUserAuthToken should overwrite an existing header", t, func() {
		req := requestWithHeader(UserAuthTokenHeader, testHeader1)

		err := SetUserAuthToken(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(UserAuthTokenHeader), ShouldEqual, testHeader2)
	})

	Convey("SetUserAuthToken should set header if it does not already exist", t, func() {
		req := requestWithHeader(UserAuthTokenHeader, "")

		err := SetUserAuthToken(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(UserAuthTokenHeader), ShouldEqual, testHeader1)
	})
}

func TestGetUserAuthToken(t *testing.T) {
	Convey("GetUserAuthToken should return expected error if request is nil", t, func() {
		actual, err := GetUserAuthToken(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetUserAuthToken should return ErrHeaderNotFound if the collection ID request header is not found", t, func() {
		req := requestWithHeader(UserAuthTokenHeader, "")

		actual, err := GetUserAuthToken(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("GetUserAuthToken should return header value if present", t, func() {
		req := requestWithHeader(UserAuthTokenHeader, testHeader1)

		actual, err := GetUserAuthToken(req)

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
