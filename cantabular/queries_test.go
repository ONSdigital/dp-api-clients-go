package cantabular

import (
	"io"

	. "github.com/smartystreets/goconvey/convey"
)

func validateQuery(body io.Reader, query string, data QueryData) {
	buf, err := data.Encode(query)
	So(err, ShouldBeNil)
	b, err := io.ReadAll(body)
	So(err, ShouldBeNil)
	So(string(b), ShouldResemble, buf.String())
}
