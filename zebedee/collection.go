package zebedee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// GetCollection returns a Collection populated with data from a zebedee response. If an error occurs, it is returned.
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

// CreateCollection creates a collection in zebedee and returns it.
func (c *Client) CreateCollection(ctx context.Context, userAccessToken string, collection Collection) (Collection, error) {
	reqURL := "/collection"

	payload, err := json.Marshal(collection)
	if err != nil {
		return Collection{}, errors.Wrap(err, "error while attempting to marshall collection")
	}

	b, _, err := c.post(ctx, userAccessToken, reqURL, payload)
	if err != nil {
		return Collection{}, err
	}

	var createdCollection Collection

	err = json.Unmarshal(b, &createdCollection)
	if err != nil {
		return Collection{}, err
	}

	return createdCollection, nil
}

// SaveContentToCollection saves the provided json content to a collection in zebedee
func (c *Client) SaveContentToCollection(ctx context.Context, userAccessToken, collectionID, pagePath string, content interface{}) error {
	reqURL := fmt.Sprintf("/content/%s?uri=%s/data.json", collectionID, pagePath)

	payload, err := json.Marshal(content)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall content")
	}

	_, _, err = c.post(ctx, userAccessToken, reqURL, payload)
	if err != nil {
		return err
	}

	return nil
}
