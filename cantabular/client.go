package cantabular

import (
	"net/http"
	"io/ioutil"
	"errors"
	"context"
	"net/url"
	"fmt"
	"encoding/json"

	"github.com/ONSdigital/log.go/v2/log"
)

// Client is the client for interacting with the Cantabular API
type Client struct{
	ua   httpClient
	host string
}

// NewClient returns a new Client
func NewClient(ua httpClient, cfg Config) *Client{
	return &Client{
		ua:   ua,
		host: cfg.Host,
	}
}

// get makes a get request to the given url and returns the response
func (c *Client) httpGet(ctx context.Context, path string) (*http.Response, error) {
	URL, err := url.Parse(path)
	if err != nil {
		return nil, &Error{
			err: fmt.Errorf("failed to parse url: %s", err),
			statusCode: http.StatusBadRequest,
			logData: log.Data{
				"url": path,
			},
		}
	}

	path = URL.String()

	resp, err := c.ua.Get(ctx, path)
	if err != nil {
		return nil, &Error{
			err: fmt.Errorf("failed to make request: %w", err),
			statusCode: http.StatusInternalServerError,
			logData: log.Data{
				"url": path,
			},
		}
	}

	return resp, nil
}

// errorResponse handles dealing with an error response from Cantabular
func (c *Client) errorResponse(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return &Error{
			err: fmt.Errorf("failed to read error response body: %s", err),
			statusCode: res.StatusCode,
		}
	}

	if len(b) == 0{
		b = []byte("[response body empty]")
	}

	var resp ErrorResponse

	if err := json.Unmarshal(b, &resp); err != nil{
		return &Error{
			err: fmt.Errorf("failed to unmarshal error response body: %s", err),
			statusCode: res.StatusCode,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	return &Error{
		err: errors.New(resp.Message),
		statusCode: res.StatusCode,
	}
}
