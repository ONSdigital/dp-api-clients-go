package upload

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
	"github.com/ONSdigital/log.go/v2/log"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
)

const (
	service     = "upload-api"
	chunkSize   = 5 * 1024 * 1024
	maxChunks   = 10000
	MaxFileSize = chunkSize * maxChunks
)

var (
	ErrFileTooLarge = errors.New(fmt.Sprintf("file too large, max file size: %d MB", MaxFileSize>>20))
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

func (c *Client) Upload(ctx context.Context, fileContent io.ReadCloser, metadata Metadata) error {
	if err := c.validateMetadata(metadata); err != nil {
		return err
	}

	totalChunks := c.calculateTotalChunks(metadata)

	for i := 1; i <= totalChunks; i++ {
		chunkContext := ChunkContext{i, totalChunks}
		reqBody, contentType, err := c.generateRequestBody(ctx, chunkContext, fileContent, metadata)
		if err != nil {
			return err
		}

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/upload-new", c.hcCli.URL), reqBody)
		req.Header.Set("Content-Type", contentType)

		resp, err := dphttp.DefaultClient.Do(ctx, req)
		if err != nil {
			log.Error(ctx, "failed request", err, log.Data{"request": req})
			return err
		}
		statusCode := resp.StatusCode

		if unsuccessfulRequest(statusCode) {
			switch statusCode {
			case http.StatusInternalServerError,
				http.StatusBadRequest,
				http.StatusUnauthorized,
				http.StatusForbidden,
				http.StatusNotFound:
				je := dperrors.JsonErrors{}
				json.NewDecoder(resp.Body).Decode(&je)

				return je.ToNativeError()
			default:
				body, _ := ioutil.ReadAll(resp.Body)
				return errors.New(string(body))
			}
		}
	}

	return nil
}

func unsuccessfulRequest(statusCode int) bool {
	return statusCode != http.StatusOK && statusCode != http.StatusCreated
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

func (c *Client) validateMetadata(metadata Metadata) error {
	if metadata.FileSizeBytes > MaxFileSize {
		return ErrFileTooLarge
	}

	return nil
}

func (c *Client) chunkReader(ctx context.Context, fileContent io.ReadCloser) (io.Reader, int, error) {
	readBuff := make([]byte, chunkSize)
	bytesRead, err := fileContent.Read(readBuff)

	if err != nil {
		log.Error(ctx, "file content read error", err)
		return nil, 0, err
	}

	var outBuff []byte

	if bytesRead == chunkSize {
		outBuff = readBuff
	} else if bytesRead < chunkSize {
		outBuff = readBuff[:bytesRead]
	}

	return bytes.NewReader(outBuff), len(outBuff), nil
}

func (c *Client) generateRequestBody(ctx context.Context, chunkContext ChunkContext, fileContent io.ReadCloser, metadata Metadata) (*bytes.Buffer, string, error) {
	reqBuff := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBuff)
	defer formWriter.Close()

	contentChunk, contentChunkLength, err := c.chunkReader(ctx, fileContent)
	if err != nil {
		return nil, "", err
	}

	c.writeMetadataFormFields(formWriter, metadata, chunkContext)
	_, err = c.writeFileFormField(formWriter, metadata, contentChunk, contentChunkLength)
	if err != nil {
		log.Error(ctx, "error writing form file content to request buffer", err)

		return nil, "", err
	}

	return reqBuff, formWriter.FormDataContentType(), nil
}

func (c *Client) calculateTotalChunks(metadata Metadata) int {
	return int(math.Ceil(float64(metadata.FileSizeBytes) / chunkSize))
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
