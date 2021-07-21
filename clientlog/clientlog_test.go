package clientlog

import (
	"testing"

	"github.com/ONSdigital/log.go/v2/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetLogData(t *testing.T) {
	uri := "testing-uri"

	Convey("When optional data does not exist assume method is GET", t, func() {
		action := "retrieving some data"
		d := buildLogData(action, uri)
		So(d, ShouldNotBeNil)
		So(len(d), ShouldEqual, 3)
		So(d["action"], ShouldEqual, action)
		So(d["uri"], ShouldEqual, uri)
		So(d["method"], ShouldEqual, "GET")
	})

	Convey("When optional data does exist", t, func() {
		action := "creating new resource"
		data := log.Data{"id": "123", "method": "POST"}
		d := buildLogData(action, uri, data)
		So(d, ShouldNotBeNil)
		So(len(d), ShouldEqual, 4)
		So(d["action"], ShouldEqual, action)
		So(d["uri"], ShouldEqual, uri)
		So(d["method"], ShouldEqual, "POST")
		So(d["id"], ShouldEqual, "123")
	})
}
