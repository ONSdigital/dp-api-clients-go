package dataset

type GetVersionMetadataSelectionInput struct {
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	DatasetID        string
	Edition          string
	Version          string
	Dimensions       []string
}
