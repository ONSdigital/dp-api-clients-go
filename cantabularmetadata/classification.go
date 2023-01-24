package cantabularmetadata

import (
	"context"
	"errors"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

func (c *Client) GetDefaultClassification(ctx context.Context, req GetDefaultClassificationRequest) (*GetDefaultClassificationResponse, error) {
	res := &struct {
		Data   Data       `json:"data"`
		Errors []GQLError `json:"errors,omitempty"`
	}{}

	data := QueryData{
		Dataset:   req.Dataset,
		Variables: req.Variables,
	}

	if err := c.queryUnmarshal(ctx, QueryDefaultClassification, data, res); err != nil {
		return nil, err
	}

	if res != nil && len(res.Errors) != 0 {
		return nil, dperrors.New(
			errors.New("error(s) returned by graphQL query"),
			res.Errors[0].StatusCode(),
			log.Data{"errors": res.Errors},
		)
	}

	var resp GetDefaultClassificationResponse
	var defaultVars []string

	for _, v := range res.Data.Dataset.Vars {
		if v.Meta.DefaultClassificationFlag == "Y" {
			defaultVars = append(defaultVars, v.Name)
		}
	}

	resp.Variables = defaultVars

	return &resp, nil
}
