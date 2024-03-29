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

// These structures correspond to GraphQL query/responses to use directly
// against the cantabular metadata service endpoint.  They would need changing
// to be used against the ext-api.

// These are updated against v1.2 of the metadata schema and would need changing
// for future releases.

// MetadataQueryResult represents the full response including both the
// table response and the dataset response
type MetadataQueryResult struct {
	TableQueryResult   *MetadataTableQuery   `json:"table_query_result"`
	DatasetQueryResult *MetadataDatasetQuery `json:"dataset_query_result"`
}

// MetadataTableQuery represents the metadata schema for a table based query.
// It's used for forming a request (see the graphql tags) and parsing a
// response (json tags)
type MetadataTableQuery struct {
	Service struct {
		Tables []struct {
			Name        graphql.String   `json:"name"`
			DatasetName graphql.String   `json:"dataset_name"`
			Label       graphql.String   `json:"label"`
			Description graphql.String   `json:"description"`
			Vars        []graphql.String `json:"vars"`
			Meta        struct {
				Alternate_Geographic_Variables []graphql.String `graphql:"Alternate_Geographic_Variables" json:"alternate_geographic_variables"`
				Contact                        struct {
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

				DatasetMnemonic2011 graphql.String `graphql:"Dataset_Mnemonic_2011" json:"dataset_mnemonic2011" `
				DatasetPopulation   graphql.String `graphql:"Dataset_Population" json:"dataset_population"`
				GeographicCoverage  graphql.String `graphql:"Geographic_Coverage" json:"geographic_coverage"`
				LastUpdated         graphql.String `graphql:"Last_Updated" json:"last_updated"`
				Observation_Type    struct {
					Observation_Type_Description graphql.String `graphql:"Observation_Type_Description" json:"observation_type_description"`

					Observation_Type_Label graphql.String `graphql:"Observation_Type_Label" json:"observation_type_label"`

					Decimal_Places graphql.String `graphql:"Decimal_Places" json:"decimal_places"`

					Prefix graphql.String `graphql:"Prefix" json:"prefix"`

					Suffix graphql.String `graphql:"Suffix" json:"suffix"`

					FillTrailingSpaces graphql.String `graphql:"FillTrailingSpaces" json:"fill_trailing_spaces"`

					NegativeSign graphql.String `graphql:"NegativeSign" json:"negative_sign"`

					Observation_Type_Code graphql.String `graphql:"Observation_Type_Code" json:"observation_type_code"`
				} `graphql:"Observation_Type" json:"observation_type"`

				Publications []struct {
					PublisherName    graphql.String `graphql:"Publisher_Name" json:"publisher_name"`
					PublicationTitle graphql.String `graphql:"Publication_Title" json:"publication_title"`
					PublisherWebsite graphql.String `graphql:"Publisher_Website" json:"publisher_website"`
				} `graphql:"Publications" json:"publications"`

				RelatedDatasets []graphql.String `graphql:"Related_Datasets" json:"related_datasets"`

				StatisticalUnit struct {
					StatisticalUnit            graphql.String `graphql:"Statistical_Unit" json:"statistical_unit"`
					StatisticalUnitDescription graphql.String `graphql:"Statistical_Unit_Description" json:"statistical_unit_description"`
				} `graphql:"Statistical_Unit" json:"statistical_unit"`

				Version graphql.String `graphql:"Version" json:"version"`
			} `json:"meta"`
		} `graphql:"tables(names: $vars)" json:"tables"`
	} `graphql:"service(lang: $lang)" json:"service"`
}

// MetadataDatasetQuery represents the metadata schema for a dataset based query.
// It's used for forming a request (see the graphql tags) and parsing a
// response (json tags)
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
			Description graphql.String `json:"description"`
			Label       graphql.String `json:"label"`
			Name        graphql.String `json:"name"`
			Meta        struct {
				DefaultClassificationFlag graphql.String `graphql:"Default_Classification_Flag" json:"default_classification_flag"`
				Mnemonic2011              graphql.String `graphql:"Mnemonic_2011" json:"mnemonic_2011"`
				Version                   graphql.String `graphql:"Version" json:"version"`

				ONSVariable struct {
					ComparabilityComments graphql.String `graphql:"Comparability_Comments" json:"comparability_comments"`
					GeographicCoverage    graphql.String `graphql:"Geographic_Coverage" json:"geographic_coverage"`
					GeographicTheme       graphql.String `graphql:"Geographic_Theme" json:"geographic_theme"`
					QualityStatementText  graphql.String `graphql:"Quality_Statement_Text"  json:"quality_statement_text"`
					QualitySummaryURL     graphql.String `graphql:"Quality_Summary_URL"  json:"quality_summary_url"`
					UkComparisonComments  graphql.String `graphql:"Uk_Comparison_Comments"  json:"uk_comparison_comments"`
					VariableMnemonic      graphql.String `graphql:"Variable_Mnemonic"  json:"variable_mnemonic"`
					VariableMnemonic2011  graphql.String `graphql:"Variable_Mnemonic_2011" json:"variable_mnemonic_2011"`
					VariableTitle         graphql.String `graphql:"Variable_Title"  json:"variable_title"`

					Version graphql.String `graphql:"Version"  json:"version"`

					Questions []struct {
						QuestionCode             graphql.String `graphql:"Question_Code" json:"question_code"`
						QuestionFirstAskedInYear graphql.String `graphql:"Question_First_Asked_In_Year" json:"question_first_asked_in_year"`
						QuestionLabel            graphql.String `graphql:"Question_Label" json:"question_label"`
						ReasonForAskingQuestion  graphql.String `graphql:"Reason_For_Asking_Question" json:"reason_for_asking_question"`
						Version                  graphql.String `graphql:"Version" json:"version"`
					} `graphql:"Questions" json:"questions,omitempty"`

					StatisticalUnit struct {
						StatisticalUnit     graphql.String `graphql:"Statistical_Unit" json:"statistical_unit"`
						StatisticalUnitDesc graphql.String `graphql:"Statistical_Unit_Description" json:"statistical_unit_desc"`
					} `graphql:"Statistical_Unit" json:"statistical_unit"`

					Topic struct {
						TopicMnemonic    graphql.String `graphql:"Topic_Mnemonic" json:"topic_mnemonic"`
						TopicDescription graphql.String `graphql:"Topic_Description" json:"topic_description"`
						TopicTitle       graphql.String `graphql:"Topic_Title" json:"topic_title"`
					} `graphql:"Topic" json:"topic"`

					VariableType struct {
						VariableTypeCode        graphql.String `graphql:"Variable_Type_Code" json:"variable_type_code"`
						VariableTypeDescription graphql.String `graphql:"Variable_Type_Description" json:"variable_type_description"`
					} `graphql:"Variable_Type" json:"variable_type"`
				} `graphql:"ONS_Variable" json:"ONS_Variable"`

				Topics []struct {
					TopicMnemonic    graphql.String `graphql:"Topic_Mnemonic" json:"topic_mnemonic"`
					TopicDescription graphql.String `graphql:"Topic_Description" json:"topic_description"`
					TopicTitle       graphql.String `graphql:"Topic_Title" json:"topic_title"`
				} `graphql:"Topics" json:"topics"`
			} `json:"meta"`
		} `graphql:"vars(names: $vars)" json:"vars"`
	} `graphql:"dataset(name: $ds, lang: $lang)" json:"dataset"`
}

// MetadataTableQueryRequest represents the params for a GraphQL table request
type MetadataTableQueryRequest struct {
	Lang      string   `json:"lang"`
	Variables []string `json:"variables"`
}

// MetadataDatasetQueryRequest represents the params for a GraphQL dataset request
type MetadataDatasetQueryRequest struct {
	Dataset   string   `json:"dataset"`
	Lang      string   `json:"lang"`
	Variables []string `json:"variables"`
}

// MetadataTableQuery takes variables (typically a ONS dataset id) and language
// params and returns a struct representing the metadata server response for a
// table query which typically contain variables (dimensions) and Cantabular
// dataset id
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

// MetadataDatasetQuery takes cantabular dataset id, language and variables
// (dimensions) params and returns a struct representing the metadata server
// response for a dataset query
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
