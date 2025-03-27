package dataset

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetadata(t *testing.T) {
	Convey("Given a json encoded Metadata struct as returned from the dp-dataset-api", t, func() {
		mj, ex := setDatasetAPIMetadata()

		Convey("When the json is unmarshalled into a dp-clients-so Metadata struct", func() {
			var mu Metadata
			err := json.Unmarshal(mj, &mu)
			So(err, ShouldBeNil)

			Convey("the dp-clients-go Metadata's Version Links is populated with the dp-dataset-api Metadata's Links section (and dp-clients-go Metadata's DatasetDetails Links is not populated)", func() {
				So(mu, ShouldResemble, ex)
			})
		})
	})
}

func setDatasetAPIMetadata() (json.RawMessage, Metadata) {
	incoming := []byte(`
{
  "contacts":[
    {
      "email":"bob@test.com",
      "name":"Bob",
      "telephone":"01657923723"
    }
  ],
  "description":"description",
  "keywords":[
    "keyword_1",
    "keyword_2"
  ],
  "latest_changes":[
    {
      "description":"change description",
      "name":"change name",
      "type":"change type"
    }
  ],
  "links":{
    "self":{
      "href":"/dataset/metadata"
    },
    "version":{
      "href":"/dataset/D1/editions/E1/versions/V1",
      "id":"V1"
    },
    "website_version":{
      "href":"/website-version"
    }
  },
  "national_statistic":true,
  "release_date":"release date",
  "title":"title",
  "headers":[
    "csv header 1",
    "csv header 2"
  ],
  "dataset_links":{
    "latest_version":{
      "href":"/dataset/D1/editions/E1/versions/V1",
      "id":"V1"
    },
    "self":{
      "href":"/dataset/UUID",
      "id":"UUID"
    }
  }
}
`)

	expected := Metadata{
		ReleaseDate: "release date",
		CSVHeader:   []string{"csv header 1", "csv header 2"},
		LatestChanges: []Change{
			{
				Description: "change description",
				Name:        "change name",
				Type:        "change type",
			},
		},
		Links: Links{
			Self: Link{
				URL: "/dataset/metadata",
			},
			Version: Link{
				URL: "/dataset/D1/editions/E1/versions/V1",
				ID:  "V1",
			},
		},
		DatasetDetails: DatasetDetails{
			Title:             "title",
			Description:       "description",
			Keywords:          &[]string{"keyword_1", "keyword_2"},
			NationalStatistic: true,
			Contacts: &[]Contact{
				{
					Name:      "Bob",
					Email:     "bob@test.com",
					Telephone: "01657923723",
				},
			},
		},
		DatasetLinks: Links{
			Self: Link{
				URL: "/dataset/UUID",
				ID:  "UUID",
			},
			LatestVersion: Link{
				URL: "/dataset/D1/editions/E1/versions/V1",
				ID:  "V1",
			},
		},
	}

	return incoming, expected
}
