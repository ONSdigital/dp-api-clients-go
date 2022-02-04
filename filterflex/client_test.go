package filterflex_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/filterflex"

	. "github.com/smartystreets/goconvey/convey"
)

func TestForwardRequest(t *testing.T){
	Convey("Given a client intialised to a mock filterFlexAPI that returns details of the incoming request", t, func() {

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			if err != nil{
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if len(b) == 0{
				b = []byte("{nil}")
			}

			fmt.Fprintf(w, "Request for: %s %s Body: %s Header: %s", r.Method, r.URL.Path, string(b), r.Header.Get("Expected"))
		}))

		defer svr.Close()

		client := filterflex.New(filterflex.Config{
			HostURL: svr.URL,
		})

		Convey("when ForwardRequest is called with a GET request with a nil body", func() {
			req := httptest.NewRequest(http.MethodGet, "/foo", nil)
			req.Header.Set("Expected", "Value")

			resp, err := client.ForwardRequest(req)
			So(err, ShouldBeNil)

			b, err := io.ReadAll(resp.Body)
			So(err,ShouldBeNil)

			expected := "Request for: GET /foo Body: {nil} Header: Value"
			So(string(b), ShouldResemble, expected)
		})

		Convey("when ForwardRequest is called with a POST request with a body", func() {
			req := httptest.NewRequest(http.MethodPost, "/bar", bytes.NewReader([]byte("I am body")))
			req.Header.Set("Expected", "OtherValue")

			resp, err := client.ForwardRequest(req)
			So(err, ShouldBeNil)

			b, err := io.ReadAll(resp.Body)
			So(err,ShouldBeNil)

			expected := "Request for: POST /bar Body: I am body Header: OtherValue"
			So(string(b), ShouldResemble, expected)
		})
	})
}
