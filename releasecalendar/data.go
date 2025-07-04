package releasecalendar

// Release represents a release
type Release struct {
	DateChanges               []ReleaseDateChange `json:"date_changes"`
	Description               ReleaseDescription  `json:"description"`
	Links                     []Link              `json:"links"`
	Markdown                  []string            `json:"markdown"`
	RelatedDatasets           []Link              `json:"related_datasets"`
	RelatedAPIDatasets        []Link              `json:"related_api_datasets"`
	RelatedDocuments          []Link              `json:"related_documents"`
	RelatedMethodology        []Link              `json:"related_methodology"`
	RelatedMethodologyArticle []Link              `json:"related_methodology_article"`
	URI                       string              `json:"uri"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}

// Link represents a link to a related resource
type Link struct {
	Summary string `json:"summary"`
	Title   string `json:"title"`
	URI     string `json:"uri"`
}

// ReleaseDescription represents the description of a release
type ReleaseDescription struct {
	CancellationNotice []string `json:"cancellation_notice"`
	Cancelled          bool     `json:"cancelled"`
	Contact            Contact  `json:"contact"`
	Finalised          bool     `json:"finalised"`
	MigrationLink      string   `json:"migration_link"`
	NationalStatistic  bool     `json:"national_statistic"`
	NextRelease        string   `json:"next_release"`
	ProvisionalDate    string   `json:"provisional_date"`
	Published          bool     `json:"published"`
	ReleaseDate        string   `json:"release_date"`
	Summary            string   `json:"summary"`
	Survey             string   `json:"survey"`
	Title              string   `json:"title"`
	WelshStatistic     bool     `json:"welsh_statistic"`
}

// Contact represents the contact details for the release
type Contact struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
}

func (r Release) Census() bool {
	return r.Description.Survey == "census"
}
