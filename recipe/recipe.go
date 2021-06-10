package recipe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/log"
)

// GetRecipe from an ID
func (c *Client) GetRecipe(ctx context.Context, userAuthToken, serviceAuthToken, recipeID string) (*Recipe, error) {
	uri := fmt.Sprintf("%s/recipes/%s", c.hcCli.URL, recipeID)

	res, err := c.doGetWithAuthHeaders(ctx, userAuthToken, serviceAuthToken, uri)
	if err != nil {
		return nil, &Error{
			err:        fmt.Errorf("failed to get response from Recipe API: %s", err),
			statusCode: http.StatusInternalServerError,
		}
	}

	defer closeResponseBody(ctx, res)

	if res.StatusCode != http.StatusOK {
		return nil, c.errorResponse(res)
	}

	var recipe Recipe

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, &Error{
			err:        fmt.Errorf("failed to read response body: %s", err),
			statusCode: res.StatusCode,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	if len(b) == 0 {
		b = []byte("[response body empty]")
	}

	if err := json.Unmarshal(b, &recipe); err != nil {
		return nil, &Error{
			err:        fmt.Errorf("failed to unmarshal response body: %s", err),
			statusCode: http.StatusInternalServerError,
			logData: log.Data{
				"response_body": string(b),
			},
		}
	}

	return &recipe, nil
}
