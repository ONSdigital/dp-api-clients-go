package dataset

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetaData_ToString(t *testing.T) {
	Convey("Given a valid metadata object return the expected text", t, func() {
		m := setupMetadata()
		expectedMetaString := expectedData(false)

		metaString := m.ToString()
		So(metaString, ShouldEqual, expectedMetaString)
	})

	Convey("Given an empty metadata object return the expected text", t, func() {
		m := Metadata{}
		expectedMetaString := expectedData(true)
		metaString := m.ToString()

		So(metaString, ShouldEqual, expectedMetaString)
	})
}

func setupMetadata() Metadata {
	m := Metadata{
		ReleaseDate: "release date",
		LatestChanges: []Change{
			{
				Description: "change description",
				Name:        "change name",
				Type:        "change type",
			},
		},
		Downloads: map[string]Download{
			"download1": {
				URL:     "url",
				Size:    "size",
				Public:  "public",
				Private: "private",
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
				{
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
				{
					Description: "methodology description",
					URL:         "methodology url",
					Title:       "methodology title",
				},
			},
			Publications: &[]Publication{
				{
					Description: "publication description",
					URL:         "publication url",
					Title:       "publication title",
				},
			},
			RelatedDatasets: &[]RelatedDataset{
				{
					URL:   "related dataset url",
					Title: "related dataset title",
				},
			},
			CanonicalTopic: "canonicalTopicID",
			Subtopics: []string{
				"secondaryTopic1ID",
				"secondaryTopic2ID",
			},
			Survey: "census",
			RelatedContent: &[]GeneralDetails{
				{
					Description: "related content description",
					HRef:        "related content url",
					Title:       "related content title",
				},
			},
			LowestGeography: "lowest geography",
		},
	}

	return m
}

func expectedData(isEmpty bool) string {
	if isEmpty {
		return "Title: \n" +
			"Description: \n" +
			"Issued: \n" +
			"Next Release: \n" +
			"Identifier: \n" +
			"Language: English\n" +
			"Latest Changes: []\n" +
			"Periodicity: \n" +
			"Distribution:\n" +
			"Unit of measure: \n" +
			"License: \n" +
			"National Statistic: false\n" +
			"Canonical Topic: \n" +
			"Survey: \n" +
			"Lowest Geography: \n"
	}

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
		"Related Links: [{related dataset url related dataset title}]\n" +
		"Canonical Topic: canonicalTopicID\n" +
		"Subtopics: [secondaryTopic1ID secondaryTopic2ID]\n" +
		"Survey: census\n" +
		"Related Content: [{related content description related content url related content title}]\n" +
		"Lowest Geography: lowest geography\n"
}

// writeToFile, helpful function to write expected and actual outputs for syntax comparison
func writeToFile(filename, line string) error {
	connection, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	_, err = connection.WriteString(line)
	if err != nil {
		return err
	}

	return nil
}
