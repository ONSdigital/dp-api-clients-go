package dataset

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetaData_ToString(t *testing.T) {
	m := setupMetadata()
	expectedMetaString := expectedData()

	metaString := m.ToString()

	Convey("Given a valid metadata object return the expected text", t, func() {

		So(metaString, ShouldEqual, expectedMetaString)
	})
}

func setupMetadata() Metadata {
	m := Metadata{
		Version: Version{
			ReleaseDate: "release date",
			LatestChanges: []Change{
				Change{
					Description: "change description",
					Name:        "change name",
					Type:        "change type",
				},
			},
			Downloads: map[string]Download{
				"download1": Download{
					URL:     "url",
					Size:    "size",
					Public:  "public",
					Private: "private",
				},
			},
		},
		DatasetDetails: DatasetDetails{
			Title:       "title",
			Description: "description",
			Publisher: &Publisher{
				URL:  "url",
				Name: "name",
				Type: "type",
			},
			Contacts: &[]Contact{
				Contact{
					Name:      "Bob",
					Email:     "bob@test.com",
					Telephone: "01657923723",
				},
			},
			Keywords:          &[]string{"keyword_1", "keyword_2"},
			NextRelease:       "next release",
			ReleaseFrequency:  "release frequency",
			UnitOfMeasure:     "unit of measure",
			License:           "license",
			NationalStatistic: true,
			Methodologies: &[]Methodology{
				Methodology{
					Description: "methodology description",
					URL:         "methodology url",
					Title:       "methodology title",
				},
			},
			Publications: &[]Publication{
				Publication{
					Description: "publication description",
					URL:         "publication url",
					Title:       "publication title",
				},
			},
			RelatedDatasets: &[]RelatedDataset{
				RelatedDataset{
					URL:   "related dataset url",
					Title: "related dataset title",
				},
			},
		},
	}

	return m
}

func expectedData() string {
	return "Title: title\n" +
		"Description: description\n" +
		"Publisher: {url name type}\n" +
		"Issued: release date\n" +
		"Next Release: next release\n" +
		"Identifier: title\n" +
		"Keywords: [keyword_1 keyword_2]\n" +
		"Language: English\n" +
		"Contact: Bob, bob@test.com, 01657923723\n" +
		"Latest Changes: [{change description change name change type}]\n" +
		"Periodicity: release frequency\n" +
		"Distribution:\n" +
		"\tExtension: download1\n" +
		"\tSize: size\n" +
		"\tURL: url\n\n" +
		"Unit of measure: unit of measure\n" +
		"License: license\n" +
		"Methodologies: [{methodology description methodology url methodology title}]\n" +
		"National Statistic: true\n" +
		"Publications: [{publication description publication url publication title}]\n" +
		"Related Links: [{related dataset url related dataset title}]\n"
}
