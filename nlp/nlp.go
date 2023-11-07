package nlp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/config"
	"github.com/ONSdigital/dp-api-clients-go/v2/nlp/models"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	"github.com/ONSdigital/log.go/v2/log"
)

// Client represents an dphttp Client that manages endpoints and URLs for NLP-enhanced services
type Client struct {
	berlinBaseURL    string
	berlinEndpoint   string
	categoryBaseURL  string
	categoryEndpoint string
	scrubberBaseURL  string
	scrubberEndpoint string
	client           dphttp.Client
}

// New initializes a new Client configured with provided NLP service settings.
func New(nlp config.NLP) *Client {
	client := dphttp.Client{}
	return &Client{
		berlinBaseURL:    nlp.BerlinAPIURL,
		berlinEndpoint:   nlp.BerlinAPIEndpoint,
		scrubberBaseURL:  nlp.ScrubberAPIURL,
		scrubberEndpoint: nlp.ScrubberAPIEndpoint,
		categoryBaseURL:  nlp.CategoryAPIURL,
		categoryEndpoint: nlp.CategoryAPIEndpoint,
		client:           client,
	}
}

func (cli *Client) GetBerlin(ctx context.Context, query string) (*models.Berlin, error) {
	var berlin models.Berlin

	berlinURL, err := buildURL(cli.berlinBaseURL+cli.berlinEndpoint, query, "q")
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "retrieving geo-spatial-data", log.Data{"berlin url": berlinURL.String()})

	resp, err := cli.client.Get(ctx, berlinURL.String())
	if err != nil {
		return nil, fmt.Errorf("error making a get request to: %s err: %w", berlinURL.String(), err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	log.Info(ctx, "bytes for berlin", log.Data{"bytes": b})

	if resp.StatusCode != http.StatusOK {
		return &berlin, fmt.Errorf("response returned non 200 status code: %d with body: %v", resp.StatusCode, b)
	}

	if err := json.Unmarshal(b, &berlin); err != nil {
		return nil, fmt.Errorf("error unmarshaling resp body to berlin model: %w", err)
	}

	return &berlin, nil
}

func (cli *Client) GetCategory(ctx context.Context, query string) (*models.Category, error) {
	var category models.Category

	categoryURL, err := buildURL(cli.categoryBaseURL+cli.categoryEndpoint, query, "query")
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "retrieving categories", log.Data{"category url": categoryURL.String()})

	resp, err := cli.client.Get(ctx, categoryURL.String())
	if err != nil {
		return nil, fmt.Errorf("error making a get request to: %s err: %w", categoryURL.String(), err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &category, fmt.Errorf("response returned non 200 status code: %d with body: %v", resp.StatusCode, b)
	}

	if err := json.Unmarshal(b, &category); err != nil {
		return nil, fmt.Errorf("error unmarshaling resp body to category model: %w", err)
	}

	return &category, nil
}

func (cli *Client) GetScrubber(ctx context.Context, query string) (*models.Scrubber, error) {
	var scrubber models.Scrubber

	berlinURL, err := buildURL(cli.scrubberBaseURL+cli.scrubberEndpoint, query, "q")
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "retrieving OAC and SIC codes", log.Data{"scrubber url": berlinURL.String()})

	resp, err := cli.client.Get(ctx, berlinURL.String())
	if err != nil {
		return nil, fmt.Errorf("error making a get request to: %s err: %w", berlinURL.String(), err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return &scrubber, fmt.Errorf("response returned non 200 status code: %d with body: %v", resp.StatusCode, b)
	}

	if err := json.Unmarshal(b, &scrubber); err != nil {
		return nil, fmt.Errorf("error unmarshaling resp body to scrubber model: %w", err)
	}

	return &scrubber, nil
}

func buildURL(baseURL, query, queryKey string) (*url.URL, error) {
	params := url.Values{}

	params.Set(queryKey, query)

	requestURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing baseURL: %w", err)
	}

	requestURL.RawQuery = params.Encode()

	return requestURL, err
}
