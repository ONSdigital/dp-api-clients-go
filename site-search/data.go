package search

// Response represents the fields for the search results as returned by dp-search-api
type Response struct {
	Count                 int           `json:"count"`
	ContentTypes          []ContentType `json:"content_types"`
	Items                 []ContentItem `json:"items"`
	Suggestions           []string      `json:"suggestions,omitempty"`
	AdditionalSuggestions []string      `json:"additional_suggestions,omitempty"`
}

// ContentType represents the specific content type for the search results with its respective count
type ContentType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ContentItem represents each search result
type ContentItem struct {
	Description Description `json:"description"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
	Matches     *Matches    `json:"matches,omitempty"`
}

// Description represents each search result description
type Description struct {
	Contact           *Contact  `json:"contact,omitempty"`
	DatasetID         string    `json:"dataset_id,omitempty"`
	Edition           string    `json:"edition,omitempty"`
	Headline1         string    `json:"headline1,omitempty"`
	Headline2         string    `json:"headline2,omitempty"`
	Headline3         string    `json:"headline3,omitempty"`
	Keywords          *[]string `json:"keywords,omitempty"`
	LatestRelease     *bool     `json:"latest_release,omitempty"`
	Language          string    `json:"language,omitempty"`
	MetaDescription   string    `json:"meta_description,omitempty"`
	NationalStatistic *bool     `json:"national_statistic,omitempty"`
	NextRelease       string    `json:"next_release,omitempty"`
	PreUnit           string    `json:"pre_unit,omitempty"`
	ReleaseDate       string    `json:"release_date,omitempty"`
	Source            string    `json:"source,omitempty"`
	Summary           string    `json:"summary"`
	Title             string    `json:"title"`
	Unit              string    `json:"unit,omitempty"`
	Highlight         Highlight `json:"highlight,omitempty"`
}

// Highlight contains specific metadata with search keyword(s) highlighted
type Highlight struct {
	Title           string    `json:"title,omitempty"`
	Keywords        *[]string `json:"keywords,omitempty"`
	Summary         string    `json:"summary,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	DatasetID       string    `json:"dataset_id,omitempty"`
	Edition         string    `json:"edition,omitempty"`
}

// Contact represents each search result contact details
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone,omitempty"`
	Email     string `json:"email"`
}

// Matches represents each search result matches
type Matches struct {
	Description struct {
		Summary         *[]MatchDetails `json:"summary"`
		Title           *[]MatchDetails `json:"title"`
		Edition         *[]MatchDetails `json:"edition,omitempty"`
		MetaDescription *[]MatchDetails `json:"meta_description,omitempty"`
		Keywords        *[]MatchDetails `json:"keywords,omitempty"`
		DatasetID       *[]MatchDetails `json:"dataset_id,omitempty"`
	} `json:"description"`
}

// MatchDetails represents each search result matches' details
type MatchDetails struct {
	Value string `json:"value,omitempty"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Department represents response from /departments/search end point
type Department struct {
	Count int               `json:"count"`
	Took  int               `json:"took"`
	Items *[]DepartmentItem `json:"items"`
}

// DepartmentItem represents a department
type DepartmentItem struct {
	Code    string             `json:"code"`
	Name    string             `json:"name"`
	URL     string             `json:"url"`
	Matches *[]DepartmentMatch `json:"matches"`
}

// DepartmentMatch represents a department matches term
type DepartmentMatch struct {
	Terms *[]MatchDetails `json:"terms"`
}

// ReleaseResponse represents response from /search/releases endpoint
type ReleaseResponse struct {
	Took      int       `json:"took"`
	Breakdown Breakdown `json:"breakdown"`
	Releases  []Release `json:"releases"`
}

type Breakdown struct {
	Total       int `json:"total"`
	Provisional int `json:"provisional,omitempty"`
	Confirmed   int `json:"confirmed,omitempty"`
	Postponed   int `json:"postponed,omitempty"`
	Published   int `json:"published,omitempty"`
	Cancelled   int `json:"cancelled,omitempty"`
	Census      int `json:"census,omitempty"`
}

type Release struct {
	URI         string              `json:"uri"`
	DateChanges []ReleaseDateChange `json:"date_changes"`
	Description ReleaseDescription  `json:"description"`
	Highlight   *Highlight          `json:"highlight,omitempty"`
}

type ReleaseDateChange struct {
	ChangeNotice string `json:"change_notice"`
	Date         string `json:"previous_date"`
}

type ReleaseDescription struct {
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	ReleaseDate     string   `json:"release_date"`
	Published       bool     `json:"published"`
	Cancelled       bool     `json:"cancelled"`
	Finalised       bool     `json:"finalised"`
	Postponed       bool     `json:"postponed"`
	Census          bool     `json:"census"`
	Keywords        []string `json:"keywords,omitempty"`
	ProvisionalDate string   `json:"provisional_date,omitempty"`
	Language        string   `json:"language,omitempty"`
}
