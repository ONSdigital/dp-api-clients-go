package zebedee

import (
	"context"
	"io"
	"net/http"

	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
)

// Clienter defines the interface for a zebedee API client.
//
//go:generate moq -out ./mocks/client.go -pkg mocks . Clienter
type Clienter interface {
	Checker(ctx context.Context, check *health.CheckState) error

	// Generic methods
	Get(ctx context.Context, authToken, path string) ([]byte, error)
	GetWithHeaders(ctx context.Context, authToken, path string) ([]byte, http.Header, error)
	Put(ctx context.Context, authToken, path string, payload []byte) (*http.Response, error)
	Post(ctx context.Context, authToken, path string, payload []byte) ([]byte, http.Header, error)

	// Get particular content types
	GetDatasetLandingPage(ctx context.Context, authToken, collectionID, lang, path string) (DatasetLandingPage, error)
	GetBreadcrumb(ctx context.Context, authToken, collectionID, lang, uri string) ([]Breadcrumb, error)
	GetDataset(ctx context.Context, authToken, collectionID, lang, uri string) (Dataset, error)
	GetHomepageContent(ctx context.Context, authToken, collectionID, lang, path string) (HomepageContent, error)
	GetFileSize(ctx context.Context, authToken, collectionID, lang, uri string) (FileSize, error)
	GetPageTitle(ctx context.Context, authToken, collectionID, lang, uri string) (PageTitle, error)
	GetPageData(ctx context.Context, authToken, collectionID, lang, uri string) (PageData, error)
	GetPageDescription(ctx context.Context, authToken, collectionID, lang, uri string) (PageDescription, error)
	GetTimeseriesMainFigure(ctx context.Context, authToken, collectionID, lang, uri string) (TimeseriesMainFigure, error)
	GetBulletin(ctx context.Context, authToken, collectionID, lang, uri string) (Bulletin, error)
	GetRelease(ctx context.Context, authToken, collectionID, lang, uri string) (Release, error)
	GetResourceBody(ctx context.Context, authToken, collectionID, lang, uri string) ([]byte, error)
	GetResourceStream(ctx context.Context, authToken, collectionID, lang, uri string) (io.ReadCloser, error)
	GetPublishedData(ctx context.Context, uriString string) ([]byte, error)
	GetPublishedIndex(ctx context.Context, params *PublishedIndexRequestParams) (PublishedIndex, error)

	// Collection management methods
	GetCollection(ctx context.Context, authToken, collectionID string) (Collection, error)
	CreateCollection(ctx context.Context, authToken string, collection Collection) (Collection, error)
	DeleteCollection(ctx context.Context, authToken, collectionID string) error
	ApproveCollection(ctx context.Context, authToken, collectionID string) error
	PublishCollection(ctx context.Context, authToken, collectionID string) error

	// Collection content management methods
	SaveContentToCollection(ctx context.Context, authToken, collectionID, pagePath string, content interface{}) error
	CompleteCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error
	ApproveCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error
	DeleteCollectionContent(ctx context.Context, authToken, collectionID, pagePath string) error
	CheckCollectionsForURI(ctx context.Context, authToken, uri string) (collectionName string, found bool, err error)
	PutDatasetInCollection(ctx context.Context, authToken, collectionID, lang, datasetID, state string) error
	PutDatasetVersionInCollection(ctx context.Context, authToken, collectionID, lang, datasetID, edition, version, state string) error
}
