package zebedee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"

	"github.com/ONSdigital/log.go/v2/log"
)

const service = "zebedee"

// Client represents a zebedee client
type Client struct {
	hcCli *healthcheck.Client
}

// ErrInvalidZebedeeResponse is returned when zebedee does not respond
// with a valid status
type ErrInvalidZebedeeResponse struct {
	ActualCode int
	URI        string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidZebedeeResponse) Error() string {
	return fmt.Sprintf("invalid response from zebedee: %d, path: %s",
		e.ActualCode,
		e.URI,
	)
}

var _ error = ErrInvalidZebedeeResponse{}

// New creates a new Zebedee Client, set ZEBEDEE_REQUEST_TIMEOUT_SECOND
// environment variable to modify default client timeout as zebedee can often be slow
// to respond
func New(zebedeeURL string) *Client {
	timeout, err := strconv.Atoi(os.Getenv("ZEBEDEE_REQUEST_TIMEOUT_SECONDS"))
	if timeout == 0 || err != nil {
		timeout = 5
	}
	hcClient := healthcheck.NewClient(service, zebedeeURL)
	hcClient.Client.SetTimeout(time.Duration(timeout) * time.Second)

	return &Client{
		hcClient,
	}
}

// NewClientWithClienter creates a new Zebedee Client using the dp-net Clienter Http Client
// which has options to set timeout and max retries.
func NewClientWithClienter(zebedeeURL string, clienter dphttp.Clienter) *Client {
	hcClient := healthcheck.NewClientWithClienter(service, zebedeeURL, clienter)

	return &Client{
		hcClient,
	}
}

// NewWithHealthClient creates a new instance of Client,
// reusing the URL and Clienter from the provided health check client.
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls zebedee health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

// Get returns a response for the requested uri in zebedee
func (c *Client) Get(ctx context.Context, userAccessToken, path string) ([]byte, error) {
	b, _, err := c.get(ctx, userAccessToken, path)
	return b, err
}

// GetWithHeaders returns a response for the requested uri in zebedee, providing the headers too
func (c *Client) GetWithHeaders(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	return c.get(ctx, userAccessToken, path)
}

// Put updates a resource in zebedee
func (c *Client) Put(ctx context.Context, userAccessToken, path string, payload []byte) (*http.Response, error) {
	resp, err := c.put(ctx, userAccessToken, path, payload)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetDatasetLandingPage returns a DatasetLandingPage populated with data from a zebedee response. If an error
// is returned there is a chance that a partly completed DatasetLandingPage is returned
func (c *Client) GetDatasetLandingPage(ctx context.Context, userAccessToken, collectionID, lang, path string) (DatasetLandingPage, error) {

	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+path)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return DatasetLandingPage{}, err
	}

	var dlp DatasetLandingPage
	if err = json.Unmarshal(b, &dlp); err != nil {
		return dlp, err
	}

	related := [][]Related{
		dlp.RelatedDatasets,
		dlp.RelatedDocuments,
		dlp.RelatedMethodology,
		dlp.RelatedMethodologyArticle,
	}

	// Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	sem := make(chan int, 10)

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e Related, element []Related) {
				defer func() {
					<-sem
					wg.Done()
				}()
				t, _ := c.GetPageTitle(ctx, userAccessToken, collectionID, lang, e.URI)
				element[i].Title = t.Title
			}(i, e, element)
		}
	}
	wg.Wait()

	datasetDownloads := make([]Dataset, len(dlp.Datasets))

	for i := range dlp.Datasets {
		d, err := c.GetDataset(ctx, userAccessToken, collectionID, lang, dlp.Datasets[i].URI)
		if err != nil {
			log.Error(ctx, "zebedee client legacy dataset returned an error", err)
			return DatasetLandingPage{}, err
		}

		datasetDownloads[i] = d
	}

	dlp.DatasetDownloads = datasetDownloads
	return dlp, nil
}

func (c *Client) get(ctx context.Context, userAccessToken, path string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", c.hcCli.URL+path, nil)
	if err != nil {
		return nil, nil, err
	}

	dprequest.AddFlorenceHeader(req, userAccessToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, nil, ErrInvalidZebedeeResponse{resp.StatusCode, req.URL.Path}
	}

	b, err := ioutil.ReadAll(resp.Body)
	return b, resp.Header, err
}

func (c *Client) put(ctx context.Context, userAccessToken, path string, payload []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	dprequest.AddFlorenceHeader(req, userAccessToken)

	resp, err := c.hcCli.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	return resp, nil
}

// GetBreadcrumb returns a Breadcrumb
func (c *Client) GetBreadcrumb(ctx context.Context, userAccessToken, collectionID, lang, uri string) ([]Breadcrumb, error) {
	b, _, err := c.get(ctx, userAccessToken, "/parents?uri="+uri)
	if err != nil {
		return nil, err
	}

	var parentsJSON []Breadcrumb
	if err = json.Unmarshal(b, &parentsJSON); err != nil {
		return nil, err
	}

	return parentsJSON, nil
}

// GetDataset returns details about a dataset from zebedee
func (c *Client) GetDataset(ctx context.Context, userAccessToken, collectionID, lang, uri string) (Dataset, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	if err != nil {
		return Dataset{}, err
	}

	var dataset Dataset

	if err = json.Unmarshal(b, &dataset); err != nil {
		return dataset, err
	}

	return c.appendDatasetFileSizes(ctx, userAccessToken, collectionID, lang, uri, dataset)
}

func (c *Client) appendDatasetFileSizes(ctx context.Context, userAccessToken, collectionID, lang, uri string, dataset Dataset) (Dataset, error) {
	for i, download := range dataset.Downloads {
		if c.downloadStoredInZebedee(download) {
			fs, err := c.GetFileSize(ctx, userAccessToken, collectionID, lang, uri+"/"+download.File)
			if err != nil {
				return dataset, err
			}

			dataset.Downloads[i].Size = strconv.Itoa(fs.Size)
		}
	}

	for i, supplementaryFile := range dataset.SupplementaryFiles {
		if c.supplementaryFileStoredInZebedee(supplementaryFile) {
			fs, err := c.GetFileSize(ctx, userAccessToken, collectionID, lang, uri+"/"+supplementaryFile.File)
			if err != nil {
				return dataset, err
			}
			dataset.SupplementaryFiles[i].Size = strconv.Itoa(fs.Size)
		}
	}

	return dataset, nil
}

func (c *Client) downloadStoredInZebedee(download Download) bool {
	return download.URI == ""
}

func (c *Client) supplementaryFileStoredInZebedee(supplementaryFile SupplementaryFile) bool {
	return supplementaryFile.URI == ""
}

func (c *Client) GetHomepageContent(ctx context.Context, userAccessToken, collectionID, lang, path string) (HomepageContent, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+path)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return HomepageContent{}, err
	}

	var homepageContent HomepageContent
	if err = json.Unmarshal(b, &homepageContent); err != nil {
		return homepageContent, err
	}

	return homepageContent, nil
}

// GetFileSize retrieves a given filesize from zebedee
func (c *Client) GetFileSize(ctx context.Context, userAccessToken, collectionID, lang, uri string) (FileSize, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/filesize", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return FileSize{}, err
	}

	var fs FileSize
	if err = json.Unmarshal(b, &fs); err != nil {
		return fs, err
	}

	return fs, nil
}

// GetPageTitle retrieves a page title from zebedee
func (c *Client) GetPageTitle(ctx context.Context, userAccessToken, collectionID, lang, uri string) (PageTitle, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+uri+"&title")
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return PageTitle{}, err
	}

	var pt PageTitle
	if err = json.Unmarshal(b, &pt); err != nil {
		return pt, err
	}

	return pt, nil
}

// GetPageDescription retrieves a page description from zebedee
func (c *Client) GetPageDescription(ctx context.Context, userAccessToken, collectionID, lang, uri string) (PageDescription, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+uri+"&description")
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return PageDescription{}, err
	}

	var desc PageDescription
	if err = json.Unmarshal(b, &desc); err != nil {
		return desc, err
	}

	return desc, nil
}

func (c *Client) GetTimeseriesMainFigure(ctx context.Context, userAccessToken, collectionID, lang, uri string) (TimeseriesMainFigure, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	if err != nil {
		return TimeseriesMainFigure{}, err
	}

	var ts TimeseriesMainFigure
	if err = json.Unmarshal(b, &ts); err != nil {
		return ts, err
	}

	return ts, nil
}

func (c *Client) PutDatasetInCollection(ctx context.Context, userAccessToken, collectionID, lang, datasetID, state string) error {
	uri := fmt.Sprintf("%s/collections/%s/datasets/%s", c.hcCli.URL, collectionID, datasetID)

	zebedeeState := CollectionState{State: state}
	payload, err := json.Marshal(zebedeeState)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	_, err = c.put(ctx, userAccessToken, uri, payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PutDatasetVersionInCollection(ctx context.Context, userAccessToken, collectionID, lang, datasetID, edition, version, state string) error {
	uri := fmt.Sprintf("%s/collections/%s/datasets/%s/editions/%s/versions/%s", c.hcCli.URL, collectionID, datasetID, edition, version)

	zebedeeState := CollectionState{State: state}
	payload, err := json.Marshal(zebedeeState)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall version")
	}

	_, err = c.put(ctx, userAccessToken, uri, payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetCollection(ctx context.Context, userAccessToken, collectionID string) (Collection, error) {
	reqURL := fmt.Sprintf("/collectionDetails/%s", collectionID)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	if err != nil {
		return Collection{}, err
	}

	var collection Collection
	if err = json.Unmarshal(b, &collection); err != nil {
		return collection, err
	}

	return collection, nil
}

// GetBulletin retrieves a bulletin from zebedee
func (c *Client) GetBulletin(ctx context.Context, userAccessToken, collectionID, lang, uri string) (Bulletin, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return Bulletin{}, err
	}

	var bulletin Bulletin
	if err = json.Unmarshal(b, &bulletin); err != nil {
		return bulletin, err
	}

	if !bulletin.Description.LatestRelease {
		// Resolve the latest release URI
		latestUrl := fmt.Sprintf("%s/latest", bulletin.URI[:strings.LastIndex(bulletin.URI, "/")])
		t, err := c.GetPageTitle(ctx, userAccessToken, collectionID, lang, latestUrl)
		if err != nil {
			log.Error(ctx, "error finding latest release URI", err, log.Data{"url": latestUrl})
			bulletin.LatestReleaseURI = latestUrl
		} else {
			bulletin.LatestReleaseURI = t.URI
		}
	}

	related := [][]Link{
		bulletin.RelatedBulletins,
		bulletin.Links,
	}

	// Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	// We use this buffered channel to limit the number of concurrent calls we make to zebedee
	sem := make(chan int, 10)

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e Link, element []Link) {
				defer func() {
					<-sem
					wg.Done()
				}()
				t, _ := c.GetPageTitle(ctx, userAccessToken, collectionID, lang, e.URI)
				element[i].Title = t.Title
				if t.Edition != "" {
					element[i].Title += fmt.Sprintf(": %s", t.Edition)
				}
			}(i, e, element)
		}
	}
	wg.Wait()

	return bulletin, nil
}

// GetRelease retrieves a release from zebedee
func (c *Client) GetRelease(ctx context.Context, userAccessToken, collectionID, lang, uri string) (Release, error) {
	// Ensure uri starts with /
	cleanUri := filepath.Clean("/" + uri)
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/data", "uri="+cleanUri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)
	if err != nil {
		return Release{}, err
	}

	var release Release
	if err = json.Unmarshal(b, &release); err != nil {
		return release, err
	}

	related := [][]Link{
		release.RelatedDocuments,
		release.RelatedDatasets,
		release.RelatedMethodology,
		release.RelatedMethodologyArticle,
		release.Links,
	}

	// Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	// We use this buffered channel to limit the number of concurrent calls we make to zebedee
	sem := make(chan int, 10)

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e Link, element []Link) {
				defer func() {
					<-sem
					wg.Done()
				}()
				t, err := c.GetPageDescription(ctx, userAccessToken, collectionID, lang, e.URI)
				if err == nil {
					element[i].Title = t.Description.Title
					if t.Description.Edition != "" {
						element[i].Title += fmt.Sprintf(": %s", t.Description.Edition)
					}
					element[i].Summary = t.Description.Summary
				}
			}(i, e, element)
		}
	}
	wg.Wait()

	return release, nil
}

// GetResourceBody returns body of a resource e.g. JSON definition of a table
func (c *Client) GetResourceBody(ctx context.Context, userAccessToken, collectionID, lang, uri string) ([]byte, error) {
	reqURL := c.createRequestURL(ctx, collectionID, lang, "/resource", "uri="+uri)
	b, _, err := c.get(ctx, userAccessToken, reqURL)

	return b, err

}

func (c *Client) createRequestURL(ctx context.Context, collectionID, lang, path, query string) string {
	if len(collectionID) > 0 {
		path += "/" + collectionID
	}

	path += "?" + url.PathEscape(query)

	if len(lang) > 0 {
		path += "&lang=" + lang
	}

	return path
}

// GetPublishedData returns []byte
func (c *Client) GetPublishedData(ctx context.Context, uriString string) ([]byte, error) {
	reqURL := c.createRequestURL(ctx, "", "", "/publisheddata", "uri="+uriString)
	content, _, err := c.get(ctx, "", reqURL)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// GetPublishedIndex returns PublishedIndex
func (c *Client) GetPublishedIndex(ctx context.Context, params *PublishedIndexRequestParams) (PublishedIndex, error) {
	reqURL := "/publishedindex"

	if params != nil {
		reqURL = fmt.Sprintf("%s?offset=%d", reqURL, params.Offset)
		if params.Limit != nil {
			reqURL = fmt.Sprintf("%s&limit=%d", reqURL, *(params.Limit))
		}
	}

	b, _, err := c.get(ctx, "", reqURL)
	if err != nil {
		return PublishedIndex{}, err
	}

	var publishedIndex PublishedIndex
	if err = json.Unmarshal(b, &publishedIndex); err != nil {
		return publishedIndex, err
	}

	return publishedIndex, nil
}

// closeResponseBody closes the response body and logs an error if unsuccessful
func closeResponseBody(ctx context.Context, resp *http.Response) {
	if resp.Body != nil {
		if err := resp.Body.Close(); err != nil {
			log.Error(ctx, "error closing http response body", err)
		}
	}
}
