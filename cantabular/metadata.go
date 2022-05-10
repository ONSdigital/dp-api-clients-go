package cantabular

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/shurcooL/graphql"
)

type MetadataQuery struct {
	Dataset struct {
		Label       graphql.String `graphql:"label"`
		Description graphql.String `graphql:"description"`
		Meta        struct {
			Source struct {
				Contact struct {
					ContactName    graphql.String `graphql:"Contact_Name"`
					ContactEmail   graphql.String `graphql:"Contact_Email"`
					ContactPhone   graphql.String `graphql:"Contact_Phone"`
					ContactWebsite graphql.String `graphql:"Contact_Website"`
				} `graphql:"Contact"`
				Licence                    graphql.String `graphql:"Licence"`
				MethodologyLink            graphql.String `graphql:"Methodology_Link"`
				MethodologyStatement       graphql.String `graphql:"Methodology_Statement"`
				NationalStatisticCertified graphql.String `graphql:"Nationals_Statistic_Certified"`
			} `graphql:"Source"`
		} `graphql:"meta"`
		Variables struct {
			Edges []struct {
				Node struct {
					Name graphql.String
					Meta struct {
						ONSVariable struct {
							VariableDescription graphql.String   `graphql:"Variable_Description"`
							Keywords            []graphql.String `graphql:"Keywords"`

							StatisticalUnit struct {
								StatisticalUnit     graphql.String `graphql:"Statistical_Unit"`
								StatisticalUnitDesc graphql.String `graphql:"Statistical_Unit_Description"`
							} `graphql:"Statistical_Unit"`
						} `graphql:"ONS_Variable"`
					}
				}
			}
		} `graphql:"variables(names: $vars)"`
	} `graphql:"dataset(name: $ds)"`
}

type MetadataQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Variables []string `json:"variables"`
}

// MetadataQuery
func (c *Client) MetadataQuery(ctx context.Context, req MetadataQueryRequest) (*MetadataQuery, error) {
	if c.gqlClient == nil {
		return nil, dperrors.New(
			errors.New("cantabular Extended API Client not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	var dims []graphql.String
	for _, v := range req.Variables {
		dims = append(dims, graphql.String(v))
	}

	vars := map[string]interface{}{
		"ds":   graphql.String(req.Dataset),
		"vars": dims,
	}

	var fq MetadataQuery
	if err := c.gqlClient.Query(ctx, &fq, vars); err != nil {

		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return &fq, nil
}
