package cantabularmetadata_test

import (
	"io"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabularmetadata"
	. "github.com/smartystreets/goconvey/convey"
)

func validateQuery(body io.Reader, query string, data cantabularmetadata.QueryData) {
	buf, err := data.Encode(query)
	So(err, ShouldBeNil)
	b, err := io.ReadAll(body)
	So(err, ShouldBeNil)
	So(string(b), ShouldResemble, buf.String())
}
