package dataset

import (
	"bytes"
	"fmt"
	"unicode"
)

// DatasetDetails represents a response dataset model from the dataset api
type DatasetDetails struct {
	ID                string            `json:"id,omitempty"`
	CollectionID      string            `json:"collection_id,omitempty"`
	Contacts          *[]Contact        `json:"contacts,omitempty"`
	Description       string            `json:"description,omitempty"`
	Keywords          *[]string         `json:"keywords,omitempty"`
	License           string            `json:"license,omitempty"`
	Links             Links             `json:"links,omitempty"`
	Methodologies     *[]Methodology    `json:"methodologies,omitempty"`
	NationalStatistic bool              `json:"national_statistic,omitempty"`
	NextRelease       string            `json:"next_release,omitempty"`
	NomisReferenceURL string            `json:"nomis_reference_url,omitempty"`
	Publications      *[]Publication    `json:"publications,omitempty"`
	Publisher         *Publisher        `json:"publisher,omitempty"`
	QMI               Publication       `json:"qmi,omitempty"`
	RelatedDatasets   *[]RelatedDataset `json:"related_datasets,omitempty"`
	ReleaseFrequency  string            `json:"release_frequency,omitempty"`
	State             string            `json:"state,omitempty"`
	Theme             string            `json:"theme,omitempty"`
	Title             string            `json:"title,omitempty"`
	Type              string            `json:"type,omitempty"`
	UnitOfMeasure     string            `json:"unit_of_measure,omitempty"`
	URI               string            `json:"uri,omitempty"`
	UsageNotes        *[]UsageNote      `json:"usage_notes,omitempty"`
}

// Dataset represents a dataset resource
type Dataset struct {
	ID      string          `json:"id"`
	Next    *DatasetDetails `json:"next,omitempty"`
	Current *DatasetDetails `json:"current,omitempty"`
	DatasetDetails
}

// List represents an object containing a list of datasets
type List struct {
	Items      []Dataset `json:"items"`
	Count      int       `json:"count"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	TotalCount int       `json:"total_count"`
}

// VersionsList represents an object containing a list of datasets
type VersionsList struct {
	Items      []Version `json:"items"`
	Count      int       `json:"count"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	TotalCount int       `json:"total_count"`
}

// NewInstance which presents a single dataset being imported
type NewInstance struct {
	InstanceID        string               `json:"id,omitempty"`
	Links             *Links               `json:"links,omitempty"`
	State             string               `json:"state,omitempty"`
	Events            []Event              `json:"events,omitempty"`
	TotalObservations int                  `json:"total_observations,omitempty"`
	Headers           []string             `json:"headers,omitempty"`
	Dimensions        []CodeList           `json:"dimensions,omitempty"`
	LastUpdated       string               `json:"last_updated,omitempty"`
	ImportTasks       *InstanceImportTasks `json:"import_tasks"`
	Type              string               `json:"type,omitempty"`
}

// Event holds one of the event which has happened to a new Instance
type Event struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	Message       string `json:"message"`
	MessageOffset string `json:"messageOffset"`
}

// CodeList holds one of the codelists corresponding to a new Instance
type CodeList struct {
	ID          string `json:"id"`
	HRef        string `json:"href"`
	Name        string `json:"name"`
	IsHierarchy bool   `json:"is_hierarchy"`
}

// Version represents a version within a dataset
type Version struct {
	Alerts               *[]Alert             `json:"alerts"`
	CollectionID         string               `json:"collection_id"`
	Downloads            map[string]Download  `json:"downloads"`
	Edition              string               `json:"edition"`
	Dimensions           []VersionDimension   `json:"dimensions"`
	ID                   string               `json:"id"`
	InstanceID           string               `json:"instance_id"`
	LatestChanges        []Change             `json:"latest_changes"`
	Links                Links                `json:"links,omitempty"`
	ReleaseDate          string               `json:"release_date"`
	State                string               `json:"state"`
	Temporal             []Temporal           `json:"temporal"`
	Version              int                  `json:"version"`
	NumberOfObservations int64                `json:"total_observations,omitempty"`
	ImportTasks          *InstanceImportTasks `json:"import_tasks,omitempty"`
	CSVHeader            []string             `json:"headers,omitempty"`
	UsageNotes           *[]UsageNote         `json:"usage_notes,omitempty"`
}

type UpdateInstance struct {
	Alerts               *[]Alert             `json:"alerts"`
	CollectionID         string               `json:"collection_id"`
	Downloads            DownloadList         `json:"downloads"`
	Edition              string               `json:"edition"`
	Dimensions           []VersionDimension   `json:"dimensions"`
	ID                   string               `json:"id"`
	InstanceID           string               `json:"instance_id"`
	LatestChanges        []Change             `json:"latest_changes"`
	ReleaseDate          string               `json:"release_date"`
	State                string               `json:"state"`
	Temporal             []Temporal           `json:"temporal"`
	Version              int                  `json:"version"`
	NumberOfObservations int64                `json:"total_observations,omitempty"`
	ImportTasks          *InstanceImportTasks `json:"import_tasks,omitempty"`
	CSVHeader            []string             `json:"headers,omitempty"`
	Type                 string               `json:"type,omitempty"`
	IsBasedOn            *IsBasedOn           `json:"is_based_on,omitempty"`
}

// VersionDimension represents a dimension model nested in the Version model
type VersionDimension struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Links           Links  `json:"links"`
	Description     string `json:"description"`
	Label           string `json:"label"`
	URL             string `json:"href,omitempty"`
	Variable        string `json:"variable,omitempty"`
	NumberOfOptions int    `json:"number_of_options,omitempty"`
}

// InstanceImportTasks represents all of the tasks required to complete an import job.
type InstanceImportTasks struct {
	ImportObservations    *ImportObservationsTask `json:"import_observations"`
	BuildHierarchyTasks   []*BuildHierarchyTask   `json:"build_hierarchies"`
	BuildSearchIndexTasks []*BuildSearchIndexTask `json:"build_search_indexes"`
}

// ImportObservationsTask represents the task of importing instance observation data into the database.
type ImportObservationsTask struct {
	State                string `json:"state,omitempty"`
	InsertedObservations int64  `json:"total_inserted_observations,omitempty"`
}

// BuildHierarchyTask represents a task of importing a single hierarchy.
type BuildHierarchyTask struct {
	State         string `json:"state,omitempty"`
	DimensionName string `json:"dimension_name,omitempty"`
	CodeListID    string `json:"code_list_id,omitempty"`
}

// BuildSearchIndexTask represents a task of importing a single search index into search.
type BuildSearchIndexTask struct {
	State         string `json:"state,omitempty"`
	DimensionName string `json:"dimension_name,omitempty"`
}

// Instance represents an instance within a dataset
type Instance struct {
	Version
}

// stateData represents a json with a single state filed
type stateData struct {
	State string `json:"state"`
}

// Instances represent a list of Instance objects
type Instances struct {
	Items      []Instance `json:"items"`
	Count      int        `json:"count"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
	TotalCount int        `json:"total_count"`
}

// Metadata is a combination of version and dataset model fields
type Metadata struct {
	Version
	DatasetDetails
}

// DownloadList represents a list of objects of containing information on the downloadable files
type DownloadList struct {
	CSV  *Download `json:"csv,omitempty"`
	CSVW *Download `json:"csvw,omitempty"`
	XLS  *Download `json:"xls,omitempty"`
}

// Download represents a version download from the dataset api
type Download struct {
	URL     string `json:"href"`
	Size    string `json:"size"`
	Public  string `json:"public,omitempty"`
	Private string `json:"private,omitempty"`
}

// Edition represents an edition within a dataset
type Edition struct {
	Edition string `json:"edition"`
	ID      string `json:"id"`
	Links   Links  `json:"links"`
	State   string `json:"state"`
}

// Publisher represents the publisher within the dataset
type Publisher struct {
	URL  string `json:"href"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// UsageNote represents a note containing extra information associated to the resource
type UsageNote struct {
	Note  string `json:"note,omitempty"`
	Title string `json:"title,omitempty"`
}

// Links represent the Links within a dataset model
type Links struct {
	AccessRights  Link `json:"access_rights,omitempty"`
	Dataset       Link `json:"dataset,omitempty"`
	Dimensions    Link `json:"dimensions,omitempty"`
	Edition       Link `json:"edition,omitempty"`
	Editions      Link `json:"editions,omitempty"`
	LatestVersion Link `json:"latest_version,omitempty"`
	Versions      Link `json:"versions,omitempty"`
	Self          Link `json:"self,omitempty"`
	CodeList      Link `json:"code_list,omitempty"`
	Options       Link `json:"options,omitempty"`
	Version       Link `json:"version,omitempty"`
	Code          Link `json:"code,omitempty"`
	Taxonomy      Link `json:"taxonomy,omitempty"`
	Job           Link `json:"job,omitempty"`
}

// Link represents a single link within a dataset model
type Link struct {
	URL string `json:"href"`
	ID  string `json:"id,omitempty"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

// VersionDimensions represent a list of versionDimension
type VersionDimensions struct {
	Items VersionDimensionItems `json:"items"`
}

// VersionDimensionItems represents a list of Version Dimensions
type VersionDimensionItems []VersionDimension

func (d VersionDimensionItems) Len() int      { return len(d) }
func (d VersionDimensionItems) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d VersionDimensionItems) Less(i, j int) bool {
	iRunes := []rune(d[i].Name)
	jRunes := []rune(d[j].Name)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}

// Dimension represents a response model for a dimension endpoint
type Dimension struct {
	DimensionID string `json:"dimension"`
	InstanceID  string `json:"instance_id"`
	NodeID      string `json:"node_id,omitempty"`
	Label       string `json:"label"`
	Option      string `json:"option"`
	Links       Links  `json:"links"`
}

// Dimensions represents a list of dimensions
type Dimensions struct {
	Items      []Dimension `json:"items"`
	Count      int         `json:"count"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
}

// Options represents a list of options from the dataset api
type Options struct {
	Items      []Option `json:"items"`
	Count      int      `json:"count"`
	Offset     int      `json:"offset"`
	Limit      int      `json:"limit"`
	TotalCount int      `json:"total_count"`
}

// Option represents a response model for an option
type Option struct {
	DimensionID string `json:"dimension"`
	Label       string `json:"label"`
	Links       Links  `json:"links"`
	Option      string `json:"option"`
}

// OptionPost represents an option model to store in the dataset api
type OptionPost struct {
	Code     string `json:"code"`
	CodeList string `json:"code_list,omitempty"`
	Label    string `json:"label"`
	Name     string `json:"dimension"`
	Option   string `json:"option"`
	Order    *int   `json:"order,omitempty"`
}

// JobInstance represents the details necessary to update (PUT) a job instance
type JobInstance struct {
	HeaderNames          []string `json:"headers"`
	NumberOfObservations int      `json:"total_observations"`
}

// Methodology represents a methodology document returned by the dataset api
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}

// Publication represents a publication document returned by the dataset api
type Publication struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}

// RelatedDataset represents a related dataset document returned by the dataset api
type RelatedDataset struct {
	URL   string `json:"href"`
	Title string `json:"title"`
}

// Alert represents an alert returned by the dataset api
type Alert struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Change represents a change returned for a version by the dataset api
type Change struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

// IsBasedOn is a special set of json-ld metadata for Cantabular datasets
// For more information on json-ld markup see:
// https://moz.com/blog/json-ld-for-beginners
type IsBasedOn struct {
	Type string `json:"@type"`
	ID   string `json:"@id"`
}

// Temporal represents a temporal returned by the dataset api
type Temporal struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Frequency string `json:"frequency"`
}

// ToString builds a string of metadata information
func (m Metadata) ToString() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Title: %s\n", m.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n", m.Description))
	if m.Publisher != nil {
		b.WriteString(fmt.Sprintf("Publisher: %s\n", *m.Publisher))
	}
	b.WriteString(fmt.Sprintf("Issued: %s\n", m.ReleaseDate))
	b.WriteString(fmt.Sprintf("Next Release: %s\n", m.NextRelease))
	b.WriteString(fmt.Sprintf("Identifier: %s\n", m.Title))
	if m.Keywords != nil {
		b.WriteString(fmt.Sprintf("Keywords: %s\n", *m.Keywords))
	}
	b.WriteString(fmt.Sprintf("Language: %s\n", "English"))
	if m.Contacts != nil {
		contacts := *m.Contacts
		if len(contacts) > 0 {
			b.WriteString(fmt.Sprintf("Contact: %s, %s, %s\n", contacts[0].Name, contacts[0].Email, contacts[0].Telephone))
		}
	}
	if len(m.Temporal) > 0 {
		b.WriteString(fmt.Sprintf("Temporal: %s\n", m.Temporal[0].Frequency))
	}
	b.WriteString(fmt.Sprintf("Latest Changes: %s\n", m.LatestChanges))
	b.WriteString(fmt.Sprintf("Periodicity: %s\n", m.ReleaseFrequency))
	b.WriteString("Distribution:\n")
	for k, v := range m.Downloads {
		b.WriteString(fmt.Sprintf("\tExtension: %s\n", k))
		b.WriteString(fmt.Sprintf("\tSize: %s\n", v.Size))
		b.WriteString(fmt.Sprintf("\tURL: %s\n\n", v.URL))
	}
	b.WriteString(fmt.Sprintf("Unit of measure: %s\n", m.UnitOfMeasure))
	b.WriteString(fmt.Sprintf("License: %s\n", m.License))
	if m.Methodologies != nil {
		b.WriteString(fmt.Sprintf("Methodologies: %s\n", *m.Methodologies))
	}
	b.WriteString(fmt.Sprintf("National Statistic: %t\n", m.NationalStatistic))
	if m.Publications != nil {
		b.WriteString(fmt.Sprintf("Publications: %s\n", *m.Publications))
	}
	if m.RelatedDatasets != nil {
		b.WriteString(fmt.Sprintf("Related Links: %s\n", *m.RelatedDatasets))
	}

	return b.String()
}

func (m Options) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("\n\tTitle: %s\n", m.Items[0].DimensionID))
	var labels, options []string

	for _, dim := range m.Items {
		labels = append(labels, dim.Label)
		options = append(options, dim.Option)
	}

	b.WriteString(fmt.Sprintf("\tLabels: %s\n", labels))
	b.WriteString(fmt.Sprintf("\tOptions: %v\n", options))

	return b.String()
}
