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
		req := requestWithHeader("")

		err := SetCollectionID(req, "")

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDKey), ShouldBeEmpty)
	})

	Convey("should overwrite an existing header", t, func() {
		req := requestWithHeader(testHeader1)

		err := SetCollectionID(req, testHeader2)

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDKey), ShouldEqual, testHeader2)
	})

	Convey("should set header if it does not already exist", t, func() {
		req := requestWithHeader("")

		err := SetCollectionID(req, testHeader1)

		So(err, ShouldBeNil)
		So(req.Header.Get(CollectionIDKey), ShouldEqual, testHeader1)
	})
}

func TestGetCollectionID(t *testing.T) {
	Convey("should return expected error if request is nil", t, func() {
		actual, err := GetCollectionID(nil)

		So(err, ShouldResemble, errRequestNil)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return ErrHeaderNotFound if the collection ID request header is not found", t, func() {
		req := requestWithHeader("")

		actual, err := GetCollectionID(req)

		So(err, ShouldResemble, ErrHeaderNotFound)
		So(actual, ShouldBeEmpty)
	})

	Convey("should return header value if present", t, func() {
		req := requestWithHeader(testHeader1)

		actual, err := GetCollectionID(req)

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, testHeader1)
	})
}

func requestWithHeader(val string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "http://localhost:456789/schwifty", nil)
	if len(val) > 0 {
		r.Header.Set(CollectionIDKey, val)
	}
	return r
}
