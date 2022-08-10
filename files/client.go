package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	ErrFileNotFound            = errors.New("file not found on dp-files-api")
	ErrFileAlreadyInCollection = errors.New("file collection ID already set")
	ErrNoFilesInCollection     = errors.New("no file in the collection")
	ErrInvalidState            = errors.New("file is in an invalid state for this action")
	ErrNotPublishable          = errors.New("file is not set as publishable")
	ErrNotAuthorized           = errors.New("you are not authorized for this action")
	ErrServer                  = errors.New("internal server error")
	ErrUnexpectedStatus        = errors.New("unexpected response status code")
	ErrBadRequest              = errors.New("bad request")
	ErrFileAlreadyRegistered   = fmt.Errorf("%w: file already registered", ErrBadRequest)
	ErrValidationError         = fmt.Errorf("%w: validation error", ErrBadRequest)
	ErrUnknown                 = fmt.Errorf("%w: unknown error", ErrBadRequest)
)

const (
	service = "files-api"
)

type FilePatch struct {
	State        string `json:"state,omitempty"`
	ETag         string `json:"etag,omitempty"`
	CollectionID string `json:"collection_id,omitempty"`
}

// Client is an files API client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli     *healthcheck.Client
	authToken string
	Version   string
}

// NewAPIClient creates a new instance of files Client with a given image API URL
func NewAPIClient(filesAPIURL, authToken string) *Client {
	return &Client{
		healthcheck.NewClient(service, filesAPIURL),
		authToken,
		"",
	}
}

// NewWithHealthClient creates a new instances of files Client using healthcheck client
func NewWithHealthClient(hcCli *healthcheck.Client) *Client {
	return &Client{
		hcCli: healthcheck.NewClientWithClienter(service, hcCli.URL, hcCli.Client),
	}
}

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

func (c *Client) PublishCollection(ctx context.Context, collectionID string) error {
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/collection/%s", c.hcCli.URL, collectionID), nil)
	dprequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := dphttp.NewClient().Do(ctx, req)
	if err != nil {
		log.Error(ctx, "failed request", err, log.Data{"request": req})
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrNoFilesInCollection
	case http.StatusConflict:
		return ErrInvalidState
	}

	return c.handleOtherCodes(resp)
}

func (c *Client) filesRootPath() string {
	if c.Version != "" {
		return fmt.Sprintf("%s/files", c.Version)
	} else {
		return "files"
	}
}

func (c *Client) GetFile(ctx context.Context, path string, authToken string) (FileMetaData, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/%s", c.hcCli.URL, c.filesRootPath(), path), nil)
	if err != nil {
		return FileMetaData{}, err
	}

	dprequest.AddServiceTokenHeader(req, authToken)

	resp, err := dphttp.NewClient().Do(ctx, req)
	if err != nil {
		return FileMetaData{}, err
	}

	metadata := FileMetaData{}

	switch resp.StatusCode {
	case http.StatusOK:
		err = json.NewDecoder(resp.Body).Decode(&metadata)
		return metadata, err
	case http.StatusNotFound:
		return metadata, dperrors.FromBody(resp.Body)
	}

	return metadata, c.handleOtherCodes(resp)
}

func (c *Client) RegisterFile(ctx context.Context, metadata FileMetaData) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", c.hcCli.URL, c.filesRootPath()), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	dprequest.AddServiceTokenHeader(req, c.authToken)

	resp, err := dphttp.NewClient().Do(ctx, req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		jsonErrors := dperrors.JsonErrors{}
		if err := json.NewDecoder(resp.Body).Decode(&jsonErrors); err != nil {
			return fmt.Errorf("%w: %s", ErrBadRequest, err)
		}
		e := jsonErrors.Errors[0]

		switch e.Code {
		case "DuplicateFileError":
			return ErrFileAlreadyRegistered
		case "ValidationError":
			return fmt.Errorf("%w: %s", ErrValidationError, e.Description)
		default:
			return fmt.Errorf("%w: %s: %s", ErrUnknown, e.Code, e.Description)
		}
	}

	return c.handleOtherCodes(resp)
}

func (c *Client) MarkFileUploaded(ctx context.Context, path string, etag string) error {
	return c.PatchFile(ctx, path, FilePatch{
		State: "UPLOADED",
		ETag:  etag,
	})
}
func (c *Client) MarkFileDecrypted(ctx context.Context, path string, etag string) error {
	return c.PatchFile(ctx, path, FilePatch{
		State: "DECRYPTED",
		ETag:  etag,
	})
}
func (c *Client) MarkFilePublished(ctx context.Context, path string, etag string) error {
	return c.PatchFile(ctx, path, FilePatch{
		State: "PUBLISHED",
		ETag:  etag,
	})
}

func (c *Client) SetCollectionID(ctx context.Context, filepath, collectionID string) error {
	return c.PatchFile(ctx, filepath, FilePatch{
		CollectionID: collectionID,
	})
}

func (c *Client) PatchFile(ctx context.Context, path string, patch FilePatch) error {
	payload, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s/%s", c.hcCli.URL, c.filesRootPath(), path), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	dprequest.AddServiceTokenHeader(req, c.authToken)
	resp, err := dphttp.NewClient().Do(ctx, req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrFileNotFound
	case http.StatusBadRequest:
		return ErrFileAlreadyInCollection
	case http.StatusConflict:
		jsonErrors := dperrors.JsonErrors{}
		if err := json.NewDecoder(resp.Body).Decode(&jsonErrors); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidState, err)
		}
		e := jsonErrors.Errors[0]

		switch e.Code {
		case "FileNotPublishable":
			return ErrFileAlreadyRegistered
		default:
			return ErrInvalidState
		}
	}

	return c.handleOtherCodes(resp)
}

func (c *Client) handleOtherCodes(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusForbidden:
		return ErrNotAuthorized
	case http.StatusInternalServerError:
		return fmt.Errorf("%w: %s", ErrServer, dperrors.FromBody(resp.Body))
	}

	return fmt.Errorf("%w: %v", ErrUnexpectedStatus, resp.StatusCode)
}
