package files

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	healthcheck "github.com/ONSdigital/dp-api-clients-go/v2/health"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/log.go/v2/log"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	ErrFileNotFound            = errors.New("file not found on dp-files-api")
	ErrFileAlreadyInCollection = errors.New("file collection ID already set")
	ErrNoFilesInCollection     = errors.New("no file in the collection")
	ErrInvalidState            = errors.New("files in an invalid state to publish")
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
	hcCli *healthcheck.Client
}

// NewAPIClient creates a new instance of files Client with a given image API URL
func NewAPIClient(filesAPIURL string) *Client {
	return &Client{
		healthcheck.NewClient(service, filesAPIURL),
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

	resp, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrFileNotFound
	} else if resp.StatusCode == http.StatusBadRequest {
		return ErrFileAlreadyInCollection
	} else if resp.StatusCode == http.StatusInternalServerError {
		je := jsonErrors{}
		json.NewDecoder(resp.Body).Decode(&je)
		var msgs []string
		for _, e := range je.Errors {
			msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Description))
		}

		return errors.New(strings.Join(msgs, "\n"))
	} else if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)

		return errors.New(fmt.Sprintf("Exepect Error: %d - %s", resp.StatusCode, string(b)))
	}

	return nil
}

func (c *Client) PublishCollection(ctx context.Context, collectionID string) error {
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/collection/%s", c.hcCli.URL, collectionID), nil)

	resp, err := dphttp.DefaultClient.Do(ctx, req)
	if err != nil {
		log.Error(ctx, "failed request", err, log.Data{"request": req})
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNoFilesInCollection
		} else if resp.StatusCode == http.StatusConflict {
			return ErrInvalidState
		} else if resp.StatusCode == http.StatusInternalServerError {
			je := jsonErrors{}
			json.NewDecoder(resp.Body).Decode(&je)
			var msgs []string
			for _, e := range je.Errors {
				msgs = append(msgs, fmt.Sprintf("%s: %s", e.Code, e.Description))
			}
			return errors.New(strings.Join(msgs, "\n"))
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.New(string(body))
		}
	}

	return nil
}
