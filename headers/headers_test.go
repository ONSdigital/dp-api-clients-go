package headers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testHeader1 = "1234567890"
	testHeader2 = "0987654321"
)

type setHeaderTestCase struct {
	description      string
	getRequestFunc   func() *http.Request
	setHeaderFunc    func(r *http.Request) error
	getHeaderFunc    func(r *http.Request) string
	assertResultFunc func(err error, val string)
}

type getHeaderTestCase struct {
	description      string
	getRequestFunc   func() *http.Request
	getHeaderFunc    func(r *http.Request) (string, error)
	assertResultFunc func(err error, val string)
}

func TestSetCollectionID(t *testing.T) {
	cases := []setHeaderTestCase{
		{
			description: "SetCollectionID should return error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetCollectionID(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				if r != nil {
					t.Fatalf("expected nil request but was not")
				}
				return ""
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
			},
		},
		{
			description: "SetCollectionID should not add header if value is empty",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetCollectionID(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(collectionIDHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "SetCollectionID should overwrite an existing header",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(collectionIDHeader, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetCollectionID(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(collectionIDHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader2)
			},
		},
		{
			description: "SetCollectionID should set header if it does not already exist",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetCollectionID(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(collectionIDHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execSetHeaderTestCases(t, cases)
}

func TestSetUserAuthToken(t *testing.T) {
	cases := []setHeaderTestCase{
		{
			description: "SetUserAuthToken should return error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserAuthToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				if r != nil {
					t.Fatalf("expected nil request but was not")
				}
				return ""
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
			},
		},
		{
			description: "SetUserAuthToken should not add header if value is empty",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserAuthToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "SetUserAuthToken should overwrite an existing header",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(userAuthTokenHeader, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserAuthToken(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader2)
			},
		},
		{
			description: "SetUserAuthToken should set header if it does not already exist",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserAuthToken(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execSetHeaderTestCases(t, cases)
}

func TestSetServiceAuthToken(t *testing.T) {
	cases := []setHeaderTestCase{
		{
			description: "SetServiceAuthToken should return error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetServiceAuthToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				if r != nil {
					t.Fatalf("expected nil request but was not")
				}
				return ""
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
			},
		},
		{
			description: "SetServiceAuthToken should not add header if value is empty",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetServiceAuthToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(serviceAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "SetServiceAuthToken should overwrite an existing header",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(serviceAuthTokenHeader, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetServiceAuthToken(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(serviceAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, bearerPrefix+testHeader2)
			},
		},
		{
			description: "SetServiceAuthToken should set header if it does not already exist",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetServiceAuthToken(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(serviceAuthTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, bearerPrefix+testHeader1)
			},
		},
	}

	execSetHeaderTestCases(t, cases)
}

func TestSetDownloadServiceToken(t *testing.T) {
	cases := []setHeaderTestCase{
		{
			description: "SetDownloadServiceToken should return error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetDownloadServiceToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				if r != nil {
					t.Fatalf("expected nil request but was not")
				}
				return ""
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
			},
		},
		{
			description: "SetDownloadServiceToken should not add header if value is empty",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetDownloadServiceToken(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(downloadServiceTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "SetDownloadServiceToken should overwrite an existing header",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(downloadServiceTokenHeader, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetDownloadServiceToken(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(downloadServiceTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader2)
			},
		},
		{
			description: "SetDownloadServiceToken should set header if it does not already exist",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetDownloadServiceToken(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(downloadServiceTokenHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execSetHeaderTestCases(t, cases)
}

func TestSetUserIdentity(t *testing.T) {
	cases := []setHeaderTestCase{
		{
			description: "SetUserIdentity should return error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserIdentity(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				if r != nil {
					t.Fatalf("expected nil request but was not")
				}
				return ""
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
			},
		},
		{
			description: "SetUserIdentity should not add header if value is empty",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserIdentity(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userIdentityHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "SetUserIdentity should overwrite an existing header",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(userIdentityHeader, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserIdentity(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userIdentityHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader2)
			},
		},
		{
			description: "SetUserIdentity should set header if it does not already exist",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return SetUserIdentity(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(userIdentityHeader)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execSetHeaderTestCases(t, cases)
}

func TestGetCollectionID(t *testing.T) {
	cases := []getHeaderTestCase{
		{
			description: "GetCollectionID should return expected error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetCollectionID(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetCollectionID should return ErrHeaderNotFound if the collection ID request header is not found",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetCollectionID(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetCollectionID should return header value if present",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(collectionIDHeader, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetCollectionID(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execGetHeaderTestCases(t, cases)
}

func TestGetUserAuthToken(t *testing.T) {
	cases := []getHeaderTestCase{
		{
			description: "GetUserAuthToken should return expected error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetUserAuthToken should return ErrHeaderNotFound if the collection ID request header is not found",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetUserAuthToken should return header value if present",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(userAuthTokenHeader, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execGetHeaderTestCases(t, cases)
}

func TestGetServiceAuthToken(t *testing.T) {
	cases := []getHeaderTestCase{
		{
			description: "GetServiceAuthToken should return expected error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetServiceAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetServiceAuthToken should return ErrHeaderNotFound if the collection ID request header is not found",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetServiceAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetServiceAuthToken should return header value if present",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(serviceAuthTokenHeader, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetServiceAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execGetHeaderTestCases(t, cases)
}

func TestGetDownloadServiceToken(t *testing.T) {
	cases := []getHeaderTestCase{
		{
			description: "GetDownloadServiceToken should return expected error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetDownloadServiceToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetDownloadServiceToken should return ErrHeaderNotFound if the collection ID request header is not found",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetDownloadServiceToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetDownloadServiceToken should return header value if present",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(downloadServiceTokenHeader, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetDownloadServiceToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execGetHeaderTestCases(t, cases)
}

func TestGetUserIdentity(t *testing.T) {
	cases := []getHeaderTestCase{
		{
			description: "GetUserIdentity should return expected error if request is nil",
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserIdentity(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetUserIdentity should return ErrHeaderNotFound if the collection ID request header is not found",
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserIdentity(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: "GetUserIdentity should return header value if present",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(userIdentityHeader, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetUserIdentity(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}

	execGetHeaderTestCases(t, cases)
}

func execSetHeaderTestCases(t *testing.T, cases []setHeaderTestCase) {
	for i, tc := range cases {
		desc := fmt.Sprintf("%d/%d) %s", i+1, len(cases), tc.description)
		Convey(desc, t, func() {
			r := tc.getRequestFunc()
			err := tc.setHeaderFunc(r)
			val := tc.getHeaderFunc(r)
			tc.assertResultFunc(err, val)
		})
	}
}

func execGetHeaderTestCases(t *testing.T, cases []getHeaderTestCase) {
	for i, tc := range cases {
		desc := fmt.Sprintf("%d/%d) %s", i+1, len(cases), tc.description)
		Convey(desc, t, func() {
			r := tc.getRequestFunc()
			val, err := tc.getHeaderFunc(r)
			tc.assertResultFunc(err, val)
		})
	}
}

func getRequestWithHeader(key, val string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "http://localhost:456789/schwifty", nil)
	if len(val) > 0 {
		r.Header.Set(key, val)
	}
	return r
}

func getRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "http://localhost:456789/schwifty", nil)
}
