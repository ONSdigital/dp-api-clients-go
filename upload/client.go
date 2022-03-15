package upload

import (
	"bytes"
	"context"
	"fmt"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

const service = "upload-api"

type Metadata struct {
	CollectionID  *string
	FileName      string
	Path          string
	IsPublishable bool
	Title         string
	FileSizeBytes int64
	FileType      string
	License       string
	LicenseURL    string
}

// Client is an upload API client which can be used to make requests to the server.
// It extends the generic healthcheck Client structure.
type Client struct {
	hcCli *healthcheck.Client
}

// NewAPIClient creates a new instance of Upload Client with a given image API URL
func NewAPIClient(uploadAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, uploadAPIURL),
	}
}

// Checker calls image api health endpoint and returns a check object to the caller.
func (c *Client) Checker(ctx context.Context, check *health.CheckState) error {
	return c.hcCli.Checker(ctx, check)
}

func (c *Client) Upload(ctx context.Context, fileContent io.ReadCloser, metadata Metadata) error {
	buff := &bytes.Buffer{}
	formWriter := multipart.NewWriter(buff)
	if metadata.CollectionID != nil {
		formWriter.WriteField("collectionId", *metadata.CollectionID)
	}
	formWriter.WriteField("resumableFilename", metadata.FileName)
	formWriter.WriteField("path", metadata.Path)
	formWriter.WriteField("isPublishable", strconv.FormatBool(metadata.IsPublishable))
	formWriter.WriteField("title", metadata.Title)
	formWriter.WriteField("resumableTotalSize", fmt.Sprintf("%d", metadata.FileSizeBytes))
	formWriter.WriteField("resumableType", metadata.FileType)
	formWriter.WriteField("licence", metadata.License)
	formWriter.WriteField("licenceURL", metadata.LicenseURL)
	formWriter.WriteField("resumableChunkNumber", "1")
	formWriter.WriteField("resumableTotalChunks", "1")

	part, _ := formWriter.CreateFormFile("file", metadata.FileName)

	fileContentBytes := make([]byte, metadata.FileSizeBytes)
	fileContent.Read(fileContentBytes)

	part.Write(fileContentBytes)

	formWriter.Close()

	defaultClient := dphttp.DefaultClient
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/upload", c.hcCli.URL), buff)
	req.Header.Set("Content-Type", formWriter.FormDataContentType())

	defaultClient.Do(ctx, req)

	return nil
}
