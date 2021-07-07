package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const whitelistBody = "Whitelist Handler has been executed"
const defaultBody = "Default (next in chain) Handler has been executed"

func TestWhitelist(t *testing.T) {

	Convey("Given an Allowed structure with methods and a whitelisted Handler function", t, func() {

		// allowed is the struct that defines allowed methods, and tracks calls to the provided Handler function.
		allowed := Allowed{
			Methods: []string{http.MethodGet, http.MethodPost},
			Handler: func(w http.ResponseWriter, req *http.Request) {
				io.WriteString(w, whitelistBody)
			},
		}

		// nextHandler represents the next handler in the chain. I.e. the handler that would be called for non-whitelisted calls.
		nextHandler := func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, defaultBody)
		}

		// whitelistedHandler represents the http.Handler that contains the whitelist logic provided by the allowed struct.
		whitelistedHandler := Whitelist(map[string]Allowed{"/health": allowed})(http.HandlerFunc(nextHandler))

		// rr implements http.ResponseWriter, for testing purposes.
		rr := httptest.NewRecorder()

		Convey("A request against a whitelisted path and method results in the handler provided in the allowed structure being executed", func() {
			req, err := http.NewRequest(http.MethodGet, "/health", nil)
			So(err, ShouldBeNil)
			whitelistedHandler.ServeHTTP(rr, req)
			So(rr.Body.String(), ShouldEqual, whitelistBody)
		})

		Convey("A request against a whitelisted path with a non-allowed method results in the default handler being executed", func() {
			req, err := http.NewRequest(http.MethodDelete, "/health", nil)
			So(err, ShouldBeNil)
			whitelistedHandler.ServeHTTP(rr, req)
			So(rr.Body.String(), ShouldEqual, defaultBody)
		})

		Convey("A request against a non-whitelisted path results in the default handler being executed", func() {
			req, err := http.NewRequest(http.MethodGet, "/my_path", nil)
			So(err, ShouldBeNil)
			whitelistedHandler.ServeHTTP(rr, req)
			So(rr.Body.String(), ShouldEqual, defaultBody)
		})
	})

}
