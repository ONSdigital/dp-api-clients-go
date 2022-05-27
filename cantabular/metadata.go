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

type MetadataQueryResult struct {
	TableQueryResult   *MetadataTableQuery
	DatasetQueryResult *MetadataDatasetQuery
}

type MetadataTableQuery struct {
	Service struct {
		Tables []struct {
			Name        graphql.String
			Label       graphql.String
			Description graphql.String
			Vars        []graphql.String
			Meta        struct {
				Contact struct {
					ContactName    graphql.String `graphql:"Contact_Name"`
					ContactEmail   graphql.String `graphql:"Contact_Email"`
					ContactPhone   graphql.String `graphql:"Contact_Phone"`
					ContactWebsite graphql.String `graphql:"Contact_Website"`
				} `graphql:"Contact"`

				CensusReleases []struct {
					CensusReleaseDescription graphql.String `graphql:"Census_Release_Description"`
					CensusReleaseNumber      graphql.String `graphql:"Census_Release_Number"`
					ReleaseDate              graphql.String `graphql:"Release_Date"`
				} `graphql:"Census_Releases"`

				DatasetMnemonic2011        graphql.String   `graphql:"Dataset_Mnemonic_2011"`
				DatasetPopulation          graphql.String   `graphql:"Dataset_Population"`
				DisseminationSource        graphql.String   `graphql:"Dissemination_Source"`
				GeographicCoverage         graphql.String   `graphql:"Geographic_Coverage"`
				GeographicVariableMnemonic graphql.String   `graphql:"Geographic_Variable_Mnemonic"`
				LastUpdated                graphql.String   `graphql:"Last_Updated"`
				Keywords                   []graphql.String `graphql:"Keywords"`

				Publications []struct {
					PublisherName    graphql.String `graphql:"Publisher_Name"`
					PublicationTitle graphql.String `graphql:"Publication_Title"`
					PublisherWebsite graphql.String `graphql:"Publisher_Website"`
				} `graphql:"Publications"`

				RelatedDatasets  []graphql.String `graphql:"Related_Datasets"`
				ReleaseFrequency graphql.String   `graphql:"Release_Frequency"`

				StatisticalUnit struct {
					StatisticalUnit            graphql.String `graphql:"Statistical_Unit"`
					StatisticalUnitDescription graphql.String `graphql:"Statistical_Unit_Description"`
				} `graphql:"Statistical_Unit"`

				UniqueUrl graphql.String `graphql:"Unique_Url"`
				Version   graphql.String `graphql:"Version"`
			}
		} `graphql:"tables(names: $vars)"`
	}
}

type MetadataDatasetQuery struct {
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
						} `graphql:"ONS_Variable" json:"ons_variable"`
					}
				}
			}
		} `graphql:"variables(names: $vars)"`
	} `graphql:"dataset(name: $ds)"`
}

type MetadataTableQueryRequest struct {
	//	Dataset   string   `json:"dataset"`
	Variables []string `json:"variables"`
}

type MetadataDatasetQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Variables []string `json:"variables"`
}

// MetadataQuery
func (c *Client) MetadataTableQuery(ctx context.Context, req MetadataTableQueryRequest) (*MetadataTableQuery, error) {
	if c.gqlClient == nil {
		return nil, dperrors.New(
			errors.New("cantabular Extended API Client not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	// XXX
	// does req do anything
	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	var datasetIDs []graphql.String
	for _, v := range req.Variables {
		datasetIDs = append(datasetIDs, graphql.String(v))
	}

	//datasetIDs = append(datasetIDs, graphql.String("LC1117EW"))

	vars := map[string]interface{}{
		"vars": datasetIDs,
	}

	var fq MetadataTableQuery
	if err := c.gqlClient.Query(ctx, &fq, vars); err != nil {

		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return &fq, nil
}

// MetadataDatasetQuery
func (c *Client) MetadataDatasetQuery(ctx context.Context, req MetadataDatasetQueryRequest) (*MetadataDatasetQuery, error) {
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

	var fq MetadataDatasetQuery
	if err := c.gqlClient.Query(ctx, &fq, vars); err != nil {

		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return &fq, nil
}
