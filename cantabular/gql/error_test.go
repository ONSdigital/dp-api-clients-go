package gql_test

import (
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular/gql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestChecker(t *testing.T) {

	Convey("A valid status code at the begining of the error message is correctly reported", t, func() {
		err := gql.Error{Message: "404 Not Found: dataset not loaded in this server"}
		So(err.StatusCode(), ShouldEqual, http.StatusNotFound)
	})

	Convey("A valid orted", t, func() {
		err := gql.Error{Message: ""}
		So(err.StatusCode(), ShouldEqual, http.StatusBadGateway)
	})

	Convey("A valid orted", t, func() {
		err := gql.Error{Message: "678 Wrong code"}
		So(err.StatusCode(), ShouldEqual, http.StatusBadGateway)
	})

	Convey("A valid orted", t, func() {
		err := gql.Error{Message: "Some other error message"}
		So(err.StatusCode(), ShouldEqual, http.StatusBadGateway)
	})

}
