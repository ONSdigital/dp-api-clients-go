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
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	ErrFileNotFound            = errors.New("file not found on dp-files-api")
	ErrFileAlreadyInCollection = errors.New("file collection ID already set")
	ErrNoFilesInCollection     = errors.New("no file in the collection")
	ErrInvalidState            = errors.New("files in an invalid state to publish")
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

type collectionIDSet struct {
	CollectionID string `json:"collection_id"`
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

func (c *Client) SetCollectionID(ctx context.Context, filepath, collectionID string) error {
	payload, err := json.Marshal(collectionIDSet{collectionID})
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/files/%s", c.hcCli.URL, filepath), bytes.NewReader(payload))
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
	}

	return c.handleOtherCodes(resp)
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
		return c.handleBadRequestResponse(resp)
	}

	return c.handleOtherCodes(resp)
}

func (c *Client) handleBadRequestResponse(resp *http.Response) error {
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

func (c *Client) handleOtherCodes(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusForbidden:
		return ErrNotAuthorized
	case http.StatusInternalServerError:
		return fmt.Errorf("%w: %s", ErrServer, dperrors.FromBody(resp.Body))
	}

	return fmt.Errorf("%w: %v", ErrUnexpectedStatus, resp.StatusCode)
}
