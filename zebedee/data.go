package zebedee

// Dataset represents a dataset response from zebedee
type Dataset struct {
	Type               string              `json:"type"`
	URI                string              `json:"uri"`
	Description        Description         `json:"description"`
	Downloads          []Download          `json:"downloads"`
	SupplementaryFiles []SupplementaryFile `json:"supplementaryFiles"`
	Versions           []Version           `json:"versions"`
}

// Download represents download within a dataset
type Download struct {
	File string `json:"file"`
	Size string
}

// FileSize represents a file size from zebedee
type FileSize struct {
	Size int `json:"fileSize"`
}

// PageTitle represents a page title from zebedee
type PageTitle struct {
	Title   string `json:"title"`
	Edition string `json:"edition"`
	URI     string `json:"uri"`
}

// SupplementaryFile represents a SupplementaryFile within a dataset
type SupplementaryFile struct {
	Title string `json:"title"`
	File  string `json:"file"`
	Size  string
}

// Version represents a version
type Version struct {
	URI         string `json:"uri"`
	ReleaseDate string `json:"updateDate"`
	Notice      string `json:"correctionNotice"`
	Label       string `json:"label"`
}

// Breadcrumb represents a breadcrumb from zebedee
type Breadcrumb struct {
	URI         string          `json:"uri"`
	Description NodeDescription `json:"description"`
	Type        string          `json:"type"`
}

// NodeDescription represents a node description
type NodeDescription struct {
	Title string `json:"title"`
}

// DatasetLandingPage is the page model of the Zebedee response for a dataset landing page type
type DatasetLandingPage struct {
	Type                      string      `json:"type"`
	URI                       string      `json:"uri"`
	Description               Description `json:"description"`
	Section                   Section     `json:"section"`
	Datasets                  []Related   `json:"datasets"`
	RelatedLinks              []Related   `json:"links"`
	RelatedFilterableDatasets []Related   `json:"relatedFilterableDatasets"`
	RelatedDatasets           []Related   `json:"relatedDatasets"`
	RelatedDocuments          []Related   `json:"relatedDocuments"`
	RelatedMethodology        []Related   `json:"relatedMethodology"`
	RelatedMethodologyArticle []Related   `json:"relatedMethodologyArticle"`
	Alerts                    []Alert     `json:"alerts"`
	Timeseries                bool        `json:"timeseries"`
}

// Description represents a page description
type Description struct {
	Title             string   `json:"title"`
	Edition           string   `json:"edition"`
	Summary           string   `json:"summary"`
	Keywords          []string `json:"keywords"`
	MetaDescription   string   `json:"metaDescription"`
	NationalStatistic bool     `json:"nationalStatistic"`
	LatestRelease     bool     `json:"latestRelease"`
	Contact           Contact  `json:"contact"`
	ReleaseDate       string   `json:"releaseDate"`
	NextRelease       string   `json:"nextRelease"`
	DatasetID         string   `json:"datasetId"`
	Unit              string   `json:"unit"`
	PreUnit           string   `json:"preUnit"`
	Source            string   `json:"source"`
	VersionLabel      string   `json:"versionLabel"`
}

// Contact represents a contact within dataset landing page
type Contact struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
}

// Section represents a markdown section
type Section struct {
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

// Figure represents a figure (charts, tables)
type Figure struct {
	Title    string `json:"title"`
	Filename string `json:"filename"`
	Version  string `json:"version"`
	URI      string `json:"uri"`
}

// Alert represents an alert
type Alert struct {
	Date     string `json:"date"`
	Markdown string `json:"markdown"`
	Type     string `json:"type"`
}

// Related stores the Title and URI for any related data (e.g. related publications on a dataset page)
// Deprecated: use Link
type Related = Link

// Link represents any link on website
type Link struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}

// TimeseriesMainFigure represents timeseries data for main figure on the homepage
type TimeseriesMainFigure struct {
	Description      TimeseriesDescription `json:"description"`
	Years            []TimeseriesDataPoint `json:"years"`
	Quarters         []TimeseriesDataPoint `json:"quarters"`
	Months           []TimeseriesDataPoint `json:"months"`
	RelatedDocuments []Related             `json:"relatedDocuments"`
	URI              string                `json:"uri"`
}

type TimeseriesDataPoint struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type TimeseriesDescription struct {
	CDID        string `json:"cdid"`
	Unit        string `json:"unit"`
	ReleaseDate string `json:"releaseDate"`
}

// HomepageContent represents the page model of the Zebedee response for the ONS homepage
type HomepageContent struct {
	Intro           Intro               `json:"intro"`
	FeaturedContent []Featured          `json:"featuredContent"`
	AroundONS       []Featured          `json:"aroundONS"`
	ServiceMessage  string              `json:"serviceMessage"`
	URI             string              `json:"uri"`
	Type            string              `json:"type"`
	Description     HomepageDescription `json:"description"`
}

type Intro struct {
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

type Featured struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	ImageID     string `json:"image"`
}

type HomepageDescription struct {
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	Keywords        []string `json:"keywords"`
	MetaDescription string   `json:"metaDescription"`
	Unit            string   `json:"unit"`
	PreUnit         string   `json:"preUnit"`
	Source          string   `json:"source"`
}

type Collection struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Inprogress      []CollectionItem `json:"inProgress"`
	Complete        []CollectionItem `json:"complete"`
	Reviewed        []CollectionItem `json:"reviewed"`
	Datasets        []CollectionItem `json:"datasets"`
	DatasetVersions []CollectionItem `json:"datasetVersions"`
	ApprovalStatus  string           `json:"approvalStatus"`
	Type            string           `json:"type"`
}

type CollectionItem struct {
	ID           string `json:"id"`
	State        string `json:"state"`
	LastEditedBy string `json:"lastEditedBy"`
	Title        string `json:"title"`
	URI          string `json:"uri"`
	Edition      string `json:"edition,omitempty"`
	Version      string `json:"version,omitempty"`
}

type CollectionState struct {
	State string `json:"state"`
}

type Bulletin struct {
	RelatedBulletins []Link      `json:"relatedBulletins"`
	Sections         []Section   `json:"sections"`
	Accordion        []Section   `json:"accordion"`
	RelatedData      []Link      `json:"relatedData"`
	Charts           []Figure    `json:"charts"`
	Tables           []Figure    `json:"tables"`
	Images           []Figure    `json:"images"`
	Equations        []Figure    `json:"equations"`
	Links            []Link      `json:"links"`
	Type             string      `json:"type"`
	URI              string      `json:"uri"`
	Description      Description `json:"description"`
	Versions         []Version   `json:"versions"`
	Alerts           []Alert     `json:"alerts"`
	LatestReleaseURI string      `json:"latestReleaseUri"`
}
