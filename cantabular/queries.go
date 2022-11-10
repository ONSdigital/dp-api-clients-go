package cantabular

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	dperrors "github.com/ONSdigital/dp-api-clients-go/v2/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	defaultLimit = 20
)

const QueryBaseVariable = `
query ($dataset: String!, $variables: [String!]!) {
	dataset(name: $dataset) {
		variables(names: $variables) {
			edges {
				node {
					mapFrom {
						edges {
							node {
								name
								label
							}
						}
					}
				}
			}
		}
	}
}`

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
			error
		}
	}
}`

// QueryAggregatedDimensionOptions is the graphQL query to obtain static dataset dimension options
// for aggregated population types
const QueryAggregatedDimensionOptions = `
query($dataset: String!, $variables: [String!]!) {
	dataset(name: $dataset) {
		variables(names: $variables){
			edges{
				node{
					name
					label
					categories{
						edges{
							node{
								label
								code
							}
						}
					}
				}
			}
		}
	}
}`

// QueryAllDimensions is the graphQL query to obtain all dimensions (variables without categories)
const QueryAllDimensions = `
query($dataset: String!) {
	dataset(name: $dataset) {
		variables {
			edges {
				node {
					name
					mapFrom {
						edges {
							node {
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

// QueryDimensions is the graphQL query to obtain all non-geography base dimensions (variables without categories)
const QueryDimensions = `
query ($dataset: String!, $text: String!, $limit: Int!, $offset: Int) {
	dataset(name: $dataset) {
		variables(rule: false, base: true) {
			totalCount
			search(text: $text, skip: $offset, first: $limit) {
				edges {
					node {
						name
						label
						categories {
							totalCount
						}
					}
				}
			}
		}
	}
}
`

// QueryDimensionsByName is the graphQL query to obtain dimensions by name (subset of variables, without categories)
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
query($dataset: String!, $limit: Int!, $offset: Int) {
	dataset(name: $dataset) {
		variables(rule: true, skip: $offset, first: $limit) {
			totalCount
			edges {
				node {
					name
					description
					mapFrom {
						edges {
							node {
								description
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
}`

const QueryDimensionsSearch = `
query($dataset: String!, $text: String!) {
	dataset(name: $dataset) {
		variables {
			search(text: $text) {
				edges {
					node {
						name
						label
						mapFrom {
							totalCount
							edges {
								node {
									name
									label
								}
							}
						}
					}
				}
			}
		}
	}
}`

// QueryAreas is the graphQL query to search for areas and area types which match a specific string.
// This can be used to retrieve a list of all the areas for a given area type, or to search for specific
// area within all area types.
const QueryAreas = `
query ($dataset: String!, $text: String!, $category: String!, $limit: Int!, $offset: Int) {
	dataset(name: $dataset) {
	  variables(rule:true, names: [ $text ]) {
		edges {
		  node {
			name
			label
			categories {
			  totalCount
			  search(text: $category, first: $limit, skip: $offset ) {
				edges {
				  node {
					code
					label
				  }
				}
			  }
			}
		  }
		}
	  }
	}
  }
`

// QueryArea is the graphQL query to search for an area which exactly match a specific string.
const QueryArea = `
query ($dataset: String!, $text: String!, $category: String!) {
  dataset(name: $dataset) {
    variables(rule: true, names: [ $text ]) {
      edges {
	node {
	  name
	  label
	  categories(codes: [ $category ]) {
	    edges {
	      node {
		code
		label
	      }
	    }
	  }
	}
      }
    }
  }
}
`

// QueryAreasWithoutPagination is the graphQL query to search for areas and area types which match a specific string.
const QueryAreasWithoutPagination = `
query ($dataset: String!, $text: String!, $category: String!, $limit: Int!, $offset: Int) {
	dataset(name: $dataset) {
	  variables(rule:true, names: [ $text ]) {
		edges {
		  node {
			name
			label
			categories {
			  totalCount
			  search(text: $category ) {
				edges {
				  node {
					code
					label
				  }
				}
			  }
			}
		  }
		}
	  }
	}
  }
`
const QueryParents = `
query ($dataset: String!, $variables: [String!]!, $limit: Int!, $offset: Int) {
  dataset(name: $dataset) {
    variables(names: $variables){
      edges{
	node{
	  label
	  name
	  isSourceOf(first: $limit, skip: $offset){
	    totalCount
	    edges{
	      node{
		label
		name
		categories{
		  totalCount
		}
	      }
	    }
	  }
	}
      }
    }
  }
}`

const QueryCategorisations = `
query ($dataset: String!, $text: String!) {
	dataset(name: $dataset) {
	  variables(rule: false, base: false) {
		search(text: $text) {
		  edges {
			node {
			  categories {
				edges {
				  node {
					label
					code
				  }
				}
			  }
			  name
			  label
			}
		  }
		}
	  }
	}
}`

const QueryParentAreaCount = `
query ($dataset: String!, $variables: [String!]!, $filters: [Filter!]! ) {
	dataset(name: $dataset) {
		table(variables: $variables, filters: $filters) {
			dimensions {
				count
				categories {
					code
					label
				}
			}
		}
	}
}`

// QueryData holds all the possible required variables to encode any of the graphql queries defined in this file.
type QueryData struct {
	PaginationParams
	Dataset   string
	Text      string
	Variables []string
	Filters   []Filter
	Category  string
	Rule      bool
	Base      bool
}

// Filter holds the fields for the Cantabular GraphQL 'Filter' object used for specifying categories
// returned in tables
type Filter struct {
	Codes    []string `json:"codes"`
	Variable string   `json:"variable"`
}

// Encode the provided graphQL query with the data in QueryData
// returns a byte buffer with the encoded query, along with any encoding error that might happen
func (data *QueryData) Encode(query string) (bytes.Buffer, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)

	if data.Limit == 0 {
		data.Limit = defaultLimit
	}
	vars := map[string]interface{}{
		"dataset":   data.Dataset,
		"variables": data.Variables,
		"filters":   data.Filters,
		"text":      data.Text,
		"limit":     data.Limit,
		"offset":    data.Offset,
		"category":  data.Category,
		"rule":      data.Rule,
		"base":      data.Base,
	}
	if len(data.Filters) > 0 {
		vars["filters"] = data.Filters
	}

	if err := enc.Encode(map[string]interface{}{
		"query":     query,
		"variables": vars,
	}); err != nil {
		return b, fmt.Errorf("failed to encode GraphQL query: %w", err)
	}

	return b, nil
}

// queryUnmarshal uses postQuery to perform a graphQL query and then un-marshals the response body to the provided value pointer v
// This method handles the response body closing.
func (c *Client) queryUnmarshal(ctx context.Context, graphQLQuery string, data QueryData, v interface{}) error {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url":        url,
		"query":      graphQLQuery,
		"query_data": data,
	}

	res, err := c.postQuery(ctx, graphQLQuery, data)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to post query: %s", err),
			http.StatusInternalServerError,
			logData,
		)
	}
	defer closeResponseBody(ctx, res)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dperrors.New(
			fmt.Errorf("failed to read response body: %s", err),
			c.StatusCode(err),
			logData,
		)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return dperrors.New(
			fmt.Errorf("failed to unmarshal response body: %s", err),
			http.StatusInternalServerError,
			logData,
		)
	}

	return nil
}

// postQuery performs a query against the Cantabular Extended API
// using the /graphql endpoint and the http client directly
// If the call is successfull, the response body is returned
// - Important: it's the caller's responsability to close the body once it has been fully processed.
func (c *Client) postQuery(ctx context.Context, graphQLQuery string, data QueryData) (*http.Response, error) {
	url := fmt.Sprintf("%s/graphql", c.extApiHost)

	logData := log.Data{
		"url": url,
	}

	b, err := data.Encode(graphQLQuery)
	logData["query"] = b.String()
	if err != nil {
		return nil, dperrors.New(err, http.StatusInternalServerError, logData)
	}

	// Do a POST call to graphQL endpoint
	res, err := c.httpPost(ctx, url, "application/json", &b)
	if err != nil {
		return nil, dperrors.New(
			fmt.Errorf("failed to make GraphQL query: %w", err),
			c.StatusCode(err),
			logData,
		)
	}

	// Check status code and return error
	if res.StatusCode != http.StatusOK {
		closeResponseBody(ctx, res)
		return nil, c.errorResponse(url, res)
	}

	return res, nil
}
