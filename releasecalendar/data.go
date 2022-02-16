package releasecalendar

// Release represents a release
type Release struct {
	DateChanges               []ReleaseDateChange `json:"dateChanges"`
	Description               ReleaseDescription  `json:"description"`
	Links                     []Link              `json:"links"`
	Markdown                  []string            `json:"markdown"`
	RelatedDatasets           []Link              `json:"relatedDatasets"`
	RelatedDocuments          []Link              `json:"relatedDocuments"`
	RelatedMethodology        []Link              `json:"relatedMethodology"`
	RelatedMethodologyArticle []Link              `json:"relatedMethodologyArticle"`
	URI                       string              `json:"uri"`
}

// ReleaseDateChange represent a date change of a release
type ReleaseDateChange struct {
	ChangeNotice string `json:"changeNotice"`
	Date         string `json:"previousDate"`
}

// Link represents a link to a related resource
type Link struct {
	Summary string `json:"summary"`
	Title   string `json:"title"`
	URI     string `json:"uri"`
}

// ReleaseDescription represents the description of a release
type ReleaseDescription struct {
	CancellationNotice []string `json:"cancellationNotice"`
	Cancelled          bool     `json:"cancelled"`
	Contact            Contact  `json:"contact"`
	Finalised          bool     `json:"finalised"`
	NationalStatistic  bool     `json:"nationalStatistic"`
	NextRelease        string   `json:"nextRelease"`
	ProvisionalDate    string   `json:"provisionalDate"`
	Published          bool     `json:"published"`
	ReleaseDate        string   `json:"releaseDate"`
	Summary            string   `json:"summary"`
	Title              string   `json:"title"`
}

// Contact represents the contact details for the release
type Contact struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
}
