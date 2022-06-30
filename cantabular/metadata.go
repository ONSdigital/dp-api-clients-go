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
	TableQueryResult   *MetadataTableQuery   `json:"table_query_result"`
	DatasetQueryResult *MetadataDatasetQuery `json:"dataset_query_result"`
}

type MetadataTableQuery struct {
	Service struct {
		Tables []struct {
			Name        graphql.String   `json:"name"`
			Label       graphql.String   `json:"label"`
			Description graphql.String   `json:"description"`
			Vars        []graphql.String `json:"vars"`
			Meta        struct {
				Contact struct {
					ContactName    graphql.String `graphql:"Contact_Name" json:"contact_name"`
					ContactEmail   graphql.String `graphql:"Contact_Email" json:"contact_email"`
					ContactPhone   graphql.String `graphql:"Contact_Phone" json:"contact_phone"`
					ContactWebsite graphql.String `graphql:"Contact_Website" json:"contact_website" `
				} `graphql:"Contact" json:"contact"`

				CensusReleases []struct {
					CensusReleaseDescription graphql.String `graphql:"Census_Release_Description" json:"census_release_description" `
					CensusReleaseNumber      graphql.String `graphql:"Census_Release_Number" json:"census_release_number" `
					ReleaseDate              graphql.String `graphql:"Release_Date" json:"release_date" `
				} `graphql:"Census_Releases" json:"census_releases"`

				DatasetMnemonic2011        graphql.String   `graphql:"Dataset_Mnemonic_2011" json:"dataset_mnemonic2011" `
				DatasetPopulation          graphql.String   `graphql:"Dataset_Population" json:"dataset_population"`
				DisseminationSource        graphql.String   `graphql:"Dissemination_Source" json:"dissemination_source"`
				GeographicCoverage         graphql.String   `graphql:"Geographic_Coverage" json:"geographic_coverage"`
				GeographicVariableMnemonic graphql.String   `graphql:"Geographic_Variable_Mnemonic" json:"geographic_variable_mnemonic"`
				LastUpdated                graphql.String   `graphql:"Last_Updated" json:"last_updated"`
				Keywords                   []graphql.String `graphql:"Keywords" json:"keywords"`

				Publications []struct {
					PublisherName    graphql.String `graphql:"Publisher_Name" json:"publisher_name"`
					PublicationTitle graphql.String `graphql:"Publication_Title" json:"publication_title"`
					PublisherWebsite graphql.String `graphql:"Publisher_Website" json:"publisher_website"`
				} `graphql:"Publications" json:"publications"`

				RelatedDatasets  []graphql.String `graphql:"Related_Datasets" json:"related_datasets"`
				ReleaseFrequency graphql.String   `graphql:"Release_Frequency" json:"release_frequency"`

				StatisticalUnit struct {
					StatisticalUnit            graphql.String `graphql:"Statistical_Unit" json:"statistical_unit"`
					StatisticalUnitDescription graphql.String `graphql:"Statistical_Unit_Description" json:"statistical_unit_description"`
				} `graphql:"Statistical_Unit" json:"statistical_unit"`

				UniqueUrl graphql.String `graphql:"Unique_Url" json:"unique_url"`
				Version   graphql.String `graphql:"Version" json:"version"`
			} `json:"meta"`
		} `graphql:"tables(names: $vars)" json:"tables"`
	} `graphql:"service(lang: $lang)" json:"service"`
}

type MetadataDatasetQuery struct {
	Dataset struct {
		Label       graphql.String `graphql:"label" json:"label"`
		Description graphql.String `graphql:"description" json:"description"`
		Meta        struct {
			Source struct {
				Contact struct {
					ContactName    graphql.String `graphql:"Contact_Name" json:"contact_name"`
					ContactEmail   graphql.String `graphql:"Contact_Email" json:"contact_email"`
					ContactPhone   graphql.String `graphql:"Contact_Phone" json:"contact_phone"`
					ContactWebsite graphql.String `graphql:"Contact_Website" json:"contact_website"`
				} `graphql:"Contact" json:"contact"`
				Licence                    graphql.String `graphql:"Licence" json:"licence"`
				MethodologyLink            graphql.String `graphql:"Methodology_Link" json:"methodology_link"`
				MethodologyStatement       graphql.String `graphql:"Methodology_Statement" json:"methodology_statement"`
				NationalStatisticCertified graphql.String `graphql:"Nationals_Statistic_Certified" json:"national_statistic_certified"`
			} `graphql:"Source" json:"source"`
		} `graphql:"meta" json:"meta"`
		Vars []struct {
			Name graphql.String `json:"name"`
			Meta struct {
				ONSVariable struct {
					VariableDescription graphql.String   `graphql:"Variable_Description" json:"variable_description"`
					Keywords            []graphql.String `graphql:"Keywords" json:"keywords"`

					StatisticalUnit struct {
						StatisticalUnit     graphql.String `graphql:"Statistical_Unit" json:"statistical_unit"`
						StatisticalUnitDesc graphql.String `graphql:"Statistical_Unit_Description" json:"statistical_unit_desc"`
					} `graphql:"Statistical_Unit" json:"statistical_unit"`
				} `graphql:"ONS_Variable" json:"ons_variable"`
			} `json:"meta"`
		} `graphql:"vars(names: $vars)" json:"vars"`
	} `graphql:"dataset(name: $ds, lang: $lang)" json:"dataset"`
}

// params for GraphQL reqs
type MetadataTableQueryRequest struct {
	Lang      string   `json:"lang"`
	Variables []string `json:"variables"`
}

type MetadataDatasetQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Lang      string   `json:"lang"`
	Variables []string `json:"variables"`
}

// MetadataQuery
func (c *Client) MetadataTableQuery(ctx context.Context, req MetadataTableQueryRequest) (*MetadataTableQuery, error) {
	if c.gqlClient == nil {
		return nil, dperrors.New(
			errors.New("cantabular metadata client not configured"),
			http.StatusServiceUnavailable,
			nil,
		)
	}

	logData := log.Data{
		"url":     fmt.Sprintf("%s/graphql", c.extApiHost),
		"request": req,
	}

	var datasetIDs []graphql.String
	for _, v := range req.Variables {
		datasetIDs = append(datasetIDs, graphql.String(v))
	}

	vars := map[string]interface{}{
		"vars": datasetIDs,
		"lang": graphql.String(req.Lang),
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
			errors.New("cantabular metadata client not configured"),
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
		"lang": graphql.String(req.Lang),
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
