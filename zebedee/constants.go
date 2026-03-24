package zebedee

// The constants in this file are for use by consuming applications to
// set common values when working with the zebedee client, such as collection
// types and language codes.

// Collection related
const (
	EmptyCollectionId       string = ""
	CollectionTypeManual    string = "manual"
	CollectionTypeScheduled string = "scheduled"
	CollectionTypeAutomated string = "automated"
)

// Language related
const (
	EnglishLangCode string = "en"
	WelshLangCode   string = "cy"
)

func getDataFileForLang(lang string) string {
	if lang == WelshLangCode {
		return "data_cy.json"
	} else {
		return "data.json"
	}
}
