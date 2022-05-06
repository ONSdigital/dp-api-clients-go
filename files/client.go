package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"io/ioutil"
	"net/http"
)

var (
	ErrFileNotFound            = errors.New("file not found on dp-files-api")
	ErrFileAlreadyInCollection = errors.New("file collection ID already set")
	ErrNoFilesInCollection     = errors.New("no file in the collection")
	ErrInvalidState            = errors.New("files in an invalid state to publish")
	ErrNotAuthorized           = errors.New("you are not authorized for this action")
)

const (
	service = "files-api"
)

type collectionIDSet struct {
	CollectionID string `json:"collection_id"`
}

// Client is an files API client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli     *healthcheck.Client
	authToken string
}

// NewAPIClient creates a new instance of files Client with a given image API URL
func NewAPIClient(filesAPIURL, authToken string) *Client {
	return &Client{
		healthcheck.NewClient(service, filesAPIURL),
		authToken,
	}
}

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

func (c *Client) SetCollectionID(ctx context.Context, filepath, collectionID string) error {
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(collectionIDSet{collectionID})

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/files/%s", c.hcCli.URL, filepath), buf)
	req.Header.Set("Content-Type", "application/json")
	dprequest.AddServiceTokenHeader(req, c.authToken)
	resp, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	} else if resp.StatusCode == http.StatusBadRequest {
		return ErrFileAlreadyInCollection
	} else if resp.StatusCode == http.StatusForbidden {
		return ErrNotAuthorized
	} else if resp.StatusCode == http.StatusInternalServerError {
		je := dperrors.JsonErrors{}
		json.NewDecoder(resp.Body).Decode(&je)

		return je.ToNativeError()
	} else if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("Exepect Error: %d - %s", resp.StatusCode, string(b)))
	}

	return nil
}

func (c *Client) PublishCollection(ctx context.Context, collectionID string) error {
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/collection/%s", c.hcCli.URL, collectionID), nil)
	dprequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		log.Error(ctx, "failed request", err, log.Data{"request": req})
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNoFilesInCollection
		} else if resp.StatusCode == http.StatusConflict {
			return ErrInvalidState
		} else if resp.StatusCode == http.StatusForbidden {
			return ErrNotAuthorized
		} else if resp.StatusCode == http.StatusInternalServerError {
			je := dperrors.JsonErrors{}
			json.NewDecoder(resp.Body).Decode(&je)

			return je.ToNativeError()
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.New("unexpected error: " + string(body))
		}
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, path string) (FileMetaData, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/files/%s", c.hcCli.URL, path), nil)
	if err != nil {
		return FileMetaData{}, err
	}

	dprequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		return FileMetaData{}, err
	}

	return c.parseGetFileResponse(resp)
}

func (c *Client) parseGetFileResponse(resp *http.Response) (FileMetaData, error) {
	metadata := FileMetaData{}
	var err error

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&metadata)
	} else {
		err = c.parseResponseErrors(resp)
	}

	return metadata, err
}

func (c *Client) parseResponseErrors(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusNotFound,
		http.StatusInternalServerError:
		je := dperrors.JsonErrors{}
		if err := json.NewDecoder(resp.Body).Decode(&je); err != nil {
			return err
		}
		return je.ToNativeError()
	case http.StatusForbidden:
		return ErrNotAuthorized
	default:
		return dperrors.NewErrorFromUnhandledStatusCode(service, resp.StatusCode)
	}

	return nil
}
