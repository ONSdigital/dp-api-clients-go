package headers

import (
	"errors"
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

type getResponseHeaderTestCase struct {
	description      string
	getResponseFunc  func() *http.Response
	getHeaderFunc    func(r *http.Response) (string, error)
	assertResultFunc func(err error, val string)
}

func TestIsErrNotFound(t *testing.T) {
	Convey("IsErrNotFound should return false if nil", t, func() {
		So(IsErrNotFound(nil), ShouldBeFalse)
	})

	Convey("IsErrNotFound should return false if err not equal to ErrHeaderNotFound ", t, func() {
		So(IsErrNotFound(errors.New("test")), ShouldBeFalse)
	})

	Convey("IsErrNotFound should return true if err equal to ErrHeaderNotFound ", t, func() {
		So(IsErrNotFound(ErrHeaderNotFound), ShouldBeTrue)
	})
}

func TestIsNotErrNotFound(t *testing.T) {
	Convey("IsNotErrNotFound should return true if error nil", t, func() {
		So(IsNotErrNotFound(nil), ShouldBeTrue)
	})

	Convey("IsNotErrNotFound should return true if error not equal to ErrHeaderNotFound", t, func() {
		So(IsNotErrNotFound(errors.New("i am an error")), ShouldBeTrue)
	})

	Convey("IsNotErrNotFound should return false if error equal to ErrHeaderNotFound", t, func() {
		So(IsNotErrNotFound(ErrHeaderNotFound), ShouldBeFalse)
	})
}

func setterTestCases(t *testing.T, fnName, headerName string, fnUnderTest func(req *http.Request, headerValue string) error) []setHeaderTestCase {
	return []setHeaderTestCase{
		{
			description: fmt.Sprintf("%s should return error if request is nil", fnName),
			getRequestFunc: func() *http.Request {
				return nil
			},
			setHeaderFunc: func(r *http.Request) error {
				return fnUnderTest(r, "")
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
			description: fmt.Sprintf("%s should not add header if value is empty", fnName),
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return fnUnderTest(r, "")
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(headerName)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrValueEmpty)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: fmt.Sprintf("%s should overwrite an existing header", fnName),
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(headerName, testHeader1)
			},
			setHeaderFunc: func(r *http.Request) error {
				return fnUnderTest(r, testHeader2)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(headerName)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader2)
			},
		},
		{
			description: fmt.Sprintf("%s should set header if it does not already exist", fnName),
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			setHeaderFunc: func(r *http.Request) error {
				return fnUnderTest(r, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) string {
				return r.Header.Get(headerName)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}
}

func TestSetCollectionID(t *testing.T) {
	cases := setterTestCases(t, "SetCollectionID", collectionIDHeader, SetCollectionID)
	execSetHeaderTestCases(t, cases)
}

func TestSetUserAuthToken(t *testing.T) {
	cases := setterTestCases(t, "SetUserAuthToken", userAuthTokenHeader, SetUserAuthToken)
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
	cases := setterTestCases(t, "SetDownloadServiceToken", downloadServiceTokenHeader, SetDownloadServiceToken)
	execSetHeaderTestCases(t, cases)
}

func TestSetUserIdentity(t *testing.T) {
	cases := setterTestCases(t, "SetUserIdentity", userIdentityHeader, SetUserIdentity)
	execSetHeaderTestCases(t, cases)
}

func TestSetRequestID(t *testing.T) {
	cases := setterTestCases(t, "SetRequestID", requestIDHeader, SetRequestID)
	execSetHeaderTestCases(t, cases)
}

func TestSetLocaleCode(t *testing.T) {
	cases := setterTestCases(t, "SetLocaleCode", localeCodeHeader, SetLocaleCode)
	execSetHeaderTestCases(t, cases)
}

func TestSetIfMatch(t *testing.T) {
	cases := setterTestCases(t, "SetIfMatch", ifMatchHeader, SetIfMatch)
	execSetHeaderTestCases(t, cases)
}

func TestSetETag(t *testing.T) {
	cases := setterTestCases(t, "SetETag", eTagHeader, SetETag)
	execSetHeaderTestCases(t, cases)
}

func getterTestCases(t *testing.T, fnName, headerName string, fnUnderTest func(req *http.Request) (string, error)) []getHeaderTestCase {
	return []getHeaderTestCase{
		{
			description: fmt.Sprintf("%s should return expected error if request is nil", fnName),
			getRequestFunc: func() *http.Request {
				return nil
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrRequestNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: fmt.Sprintf("%s should return ErrHeaderNotFound if the collection ID request header is not found", fnName),
			getRequestFunc: func() *http.Request {
				return getRequest()
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: fmt.Sprintf("%s should return header value if present", fnName),
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(headerName, testHeader1)
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}
}

func TestGetCollectionID(t *testing.T) {
	cases := getterTestCases(t, "GetCollectionID", collectionIDHeader, GetCollectionID)
	execGetHeaderTestCases(t, cases)
}

func TestGetUserAuthToken(t *testing.T) {
	cases := getterTestCases(t, "GetUserAuthToken", userAuthTokenHeader, GetUserAuthToken)
	execGetHeaderTestCases(t, cases)
}

func TestGetServiceAuthToken(t *testing.T) {
	cases := getterTestCases(t, "GetServiceAuthToken", serviceAuthTokenHeader, GetServiceAuthToken)
	cases = append(cases,
		getHeaderTestCase{
			description: "GetServiceAuthToken should return header value if present, trimming the 'Bearer prefix'",
			getRequestFunc: func() *http.Request {
				return getRequestWithHeader(serviceAuthTokenHeader, fmt.Sprintf("%s%s", bearerPrefix, testHeader1))
			},
			getHeaderFunc: func(r *http.Request) (string, error) {
				return GetServiceAuthToken(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	)
	execGetHeaderTestCases(t, cases)
}

func TestGetDownloadServiceToken(t *testing.T) {
	cases := getterTestCases(t, "GetDownloadServiceToken", downloadServiceTokenHeader, GetDownloadServiceToken)
	execGetHeaderTestCases(t, cases)
}

func TestGetUserIdentity(t *testing.T) {
	cases := getterTestCases(t, "GetUserIdentity", userIdentityHeader, GetUserIdentity)
	execGetHeaderTestCases(t, cases)
}

func TestGetRequestID(t *testing.T) {
	cases := getterTestCases(t, "GetRequestID", requestIDHeader, GetRequestID)
	execGetHeaderTestCases(t, cases)
}

func TestGetLocaleCode(t *testing.T) {
	cases := getterTestCases(t, "GetLocaleCode", localeCodeHeader, GetLocaleCode)
	execGetHeaderTestCases(t, cases)
}

func TestGetIfMatch(t *testing.T) {
	cases := getterTestCases(t, "GetIfMatch", ifMatchHeader, GetIfMatch)
	execGetHeaderTestCases(t, cases)
}

func TestGetETag(t *testing.T) {
	cases := getterTestCases(t, "GetETag", eTagHeader, GetETag)
	execGetHeaderTestCases(t, cases)
}

func responseGetterTestCases(t *testing.T, fnName, headerName string, fnUnderTest func(resp *http.Response) (string, error)) []getResponseHeaderTestCase {
	return []getResponseHeaderTestCase{
		{
			description: fmt.Sprintf("%s should return expected error if response is nil", fnName),
			getResponseFunc: func() *http.Response {
				return nil
			},
			getHeaderFunc: func(r *http.Response) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrResponseNil)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: fmt.Sprintf("%s should return ErrHeaderNotFound if the collection ID response header is not found", fnName),
			getResponseFunc: func() *http.Response {
				return getResponse()
			},
			getHeaderFunc: func(r *http.Response) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldResemble, ErrHeaderNotFound)
				So(val, ShouldBeEmpty)
			},
		},
		{
			description: fmt.Sprintf("%s should return header value if present", fnName),
			getResponseFunc: func() *http.Response {
				return getResponseWithHeader(headerName, testHeader1)
			},
			getHeaderFunc: func(r *http.Response) (string, error) {
				return fnUnderTest(r)
			},
			assertResultFunc: func(err error, val string) {
				So(err, ShouldBeNil)
				So(val, ShouldEqual, testHeader1)
			},
		},
	}
}

func TestGetResponseETag(t *testing.T) {
	cases := responseGetterTestCases(t, "GetResponseETag", eTagHeader, GetResponseETag)
	execResponseGetHeaderTestCases(t, cases)
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

func execResponseGetHeaderTestCases(t *testing.T, cases []getResponseHeaderTestCase) {
	for i, tc := range cases {
		desc := fmt.Sprintf("%d/%d) %s", i+1, len(cases), tc.description)
		Convey(desc, t, func() {
			r := tc.getResponseFunc()
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

func getResponseWithHeader(key, val string) *http.Response {
	r := &http.Response{Header: make(http.Header)}
	if len(val) > 0 {
		r.Header.Set(key, val)
	}
	return r
}

func getResponse() *http.Response {
	return &http.Response{Header: make(http.Header)}
}
