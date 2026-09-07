package zebedee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// GetCollection returns a collection from zebedee.
func (c *Client) GetCollection(ctx context.Context, authToken, collectionID string) (Collection, error) {
	reqURL := fmt.Sprintf("/collectionDetails/%s", collectionID)
	b, _, err := c.get(ctx, authToken, reqURL)

	if err != nil {
		return Collection{}, err
	}

	var collection Collection
	if err := json.Unmarshal(b, &collection); err != nil {
		return collection, err
	}

	return collection, nil
}

// CreateCollection creates a collection in zebedee and returns it.
func (c *Client) CreateCollection(ctx context.Context, authToken string, collection Collection) (Collection, error) {
	reqURL := "/collection"

	payload, err := json.Marshal(collection)
	if err != nil {
		return Collection{}, errors.Wrap(err, "error while attempting to marshall collection")
	}

	b, _, err := c.post(ctx, authToken, reqURL, payload)
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

// DeleteCollection deletes a collection in zebedee.
func (c *Client) DeleteCollection(ctx context.Context, authToken, collectionID string) error {
	reqURL := fmt.Sprintf("/collection/%s", collectionID)

	_, _, err := c.delete(ctx, authToken, reqURL)
	if err != nil {
		return err
	}

	return nil
}

// ApproveCollection approves a collection in zebedee.
func (c *Client) ApproveCollection(ctx context.Context, authToken, collectionID string) error {
	reqURL := fmt.Sprintf("/approve/%s", collectionID)

	_, _, err := c.post(ctx, authToken, reqURL, []byte{})
	if err != nil {
		return err
	}

	return nil
}

// PublishCollection publishes a collection in zebedee.
func (c *Client) PublishCollection(ctx context.Context, authToken, collectionID string) error {
	reqURL := fmt.Sprintf("/publish/%s", collectionID)

	_, _, err := c.post(ctx, authToken, reqURL, []byte{})
	if err != nil {
		return err
	}

	return nil
}

// SaveContentToCollection saves the provided json content
// to a collection in zebedee
func (c *Client) SaveContentToCollection(ctx context.Context, authToken, collectionID, pagePath string, content interface{}) error {
	reqURL := fmt.Sprintf("/content/%s?uri=%s/data.json", collectionID, pagePath)

	payload, err := json.Marshal(content)
	if err != nil {
		return errors.Wrap(err, "error while attempting to marshall content")
	}

	_, _, err = c.post(ctx, authToken, reqURL, payload)
	if err != nil {
		return err
	}

	return nil
}

// CompleteCollectionContent marks the content
// as completed and ready for review in zebedee
func (c *Client) CompleteCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error {
	reqURL := fmt.Sprintf("/complete/%s?uri=%s/%s", collectionID, pagePath, getDataFileForLang(lang))

	_, _, err := c.post(ctx, authToken, reqURL, []byte{})
	if err != nil {
		return err
	}

	return nil
}

// ApproveCollectionContent approves the provided json content path
// in a collection in zebedee
func (c *Client) ApproveCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error {
	reqURL := fmt.Sprintf("/review/%s?uri=%s/%s", collectionID, pagePath, getDataFileForLang(lang))

	_, _, err := c.post(ctx, authToken, reqURL, []byte{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteCollectionContent deletes the content at the provided content path
// in a collection in zebedee
func (c *Client) DeleteCollectionContent(ctx context.Context, authToken, collectionID, pagePath string) error {
	reqURL := fmt.Sprintf("/page/%s?uri=%s", collectionID, pagePath)

	_, _, err := c.delete(ctx, authToken, reqURL)
	if err != nil {
		return err
	}

	return nil
}

// CheckCollectionsForURI checks whether the given uri
// is already being edited in a collection.
func (c *Client) CheckCollectionsForURI(ctx context.Context, authToken, uri string) (collectionName string, found bool, err error) {
	reqURL := "/CheckCollectionsForURI?uri=" + uri

	b, _, err := c.get(ctx, authToken, reqURL)
	if err != nil {
		return "", false, err
	}

	name := string(b)
	return name, name != "", nil
}
