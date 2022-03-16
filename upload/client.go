package upload

import (
	"bytes"
	"context"
	"fmt"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
)

const (
	service   = "upload-api"
	chunkSize = 5 * 1024 * 1024
)

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

type ChunkContext struct {
	Current int
	Total   int
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

func (c *Client) writeMetadataFormFields(formWriter *multipart.Writer, metadata Metadata, chunk ChunkContext) {
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
	formWriter.WriteField("resumableChunkNumber", fmt.Sprintf("%d", chunk.Current))
	formWriter.WriteField("resumableTotalChunks", fmt.Sprintf("%d", chunk.Total))
}

func (c *Client) Upload(ctx context.Context, fileContent io.ReadCloser, metadata Metadata) error {
	totalChunks := int(math.Ceil(float64(metadata.FileSizeBytes) / chunkSize))

	for i := 1; i <= totalChunks; i++ {
		reqBuff := &bytes.Buffer{}
		chunkContext := ChunkContext{
			Current: i,
			Total:   totalChunks,
		}
		formWriter := multipart.NewWriter(reqBuff)

		readBuff := make([]byte, chunkSize)
		bytesRead, err := fileContent.Read(readBuff)

		if err != nil {
			log.Error(ctx, "file content read error", err)
			formWriter.Close()
			return err
		}

		var outBuff []byte

		if bytesRead == chunkSize {
			outBuff = readBuff
		} else if bytesRead < chunkSize {
			outBuff = readBuff[:bytesRead]
		}

		br := bytes.NewReader(outBuff)

		c.writeMetadataFormFields(formWriter, metadata, chunkContext)
		_, err = c.writeFileFormField(formWriter, metadata, br, len(outBuff))

		if err != nil {
			log.Error(ctx, "error writing form file content to request buffer", err)
			formWriter.Close()
			return err
		}

		formWriter.Close()

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/upload", c.hcCli.URL), reqBuff)

		req.Header.Set("Content-Type", formWriter.FormDataContentType())

		_, err = dphttp.DefaultClient.Do(ctx, req)
		if err != nil {
			log.Error(ctx, "failed request", err, log.Data{"request": req})
			return err
		}
	}

	return nil
}

func (c *Client) writeFileFormField(formWriter *multipart.Writer, metadata Metadata, fileContent io.Reader, fileSizeBytes int) (int, error) {
	part, err := formWriter.CreateFormFile("file", metadata.FileName)
	if err != nil {
		return 0, err
	}

	fileContentBytes := make([]byte, fileSizeBytes)
	fileContent.Read(fileContentBytes)

	return part.Write(fileContentBytes)
}
