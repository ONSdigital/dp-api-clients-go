package cantabular

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// QueryStaticDataset is the graphQL query to obtain static dataset counts (variables with categories and counts)
const QueryStaticDataset = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				count
				variable { name label }
				categories { code label }
			}
			values
			error
		}
	}
}`

// QueryDimensionOptions is the graphQL query to obtain static dataset dimension options (variables with categories)
const QueryDimensionOptions = `
query($dataset: String!, $variables: [String!]!, $filters: [Filter!]) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				variable { name label }
				categories { code label }
			}
			values
			error
		}
	}
}`

// QueryDimensions is the graphQL query to obtain dimensions (variables without categories)
const QueryDimensions = `
query($dataset: String!) {
	dataset(name: $dataset) {
		variables {
			edges {
				node {
					name
					mapFrom {
						edges {
							node {
								filterOnly
								label
								name
							}
						}
					}
					label
					categories {
						totalCount
					}
				}
			}
		}
	}
}`

// QueryDimensions is the graphQL query to obtain dimensions by name (subset of variables, without categories)
const QueryDimensionsByName = `
query($dataset: String!, $variables: [String!]!) {
	dataset(name: $dataset) {
		variables(names: $variables) {
			edges {
				node {
					name
					mapFrom {
						edges {
							node {
								filterOnly
								label
								name
							}
						}
					}
					label
					categories {
						totalCount
					}
				}
			}
		}
	}
}`

// QueryGeographyDimensions is the graphQL query to obtain geography dimensions (subset of variables, without categories)
const QueryGeographyDimensions = `
query($dataset: String!) {
	dataset(name: $dataset) {
		ruleBase {
			name
			isSourceOf {
				edges {
					node {
						name
						mapFrom {
							edges {
								node {
									filterOnly
									label
									name
								}
							}
						}
						label
						categories{
							totalCount
						}
					}
				}
			}
		}
	}
}`

// QueryData holds the required variables to encode a graphql query
type QueryData struct {
	Dataset   string
	Variables []string
}

// encode the provided graphQL query with the data in QueryData
// returns a byte buffer with the encoded query, along with any encoding error that might happen
func (data *QueryData) encode(query string) (bytes.Buffer, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"dataset":   data.Dataset,
			"variables": data.Variables,
		},
	}); err != nil {
		return b, fmt.Errorf("failed to encode GraphQL query: %w", err)
	}
	return b, nil
}
