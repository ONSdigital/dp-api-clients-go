package cantabular

// QueryStaticDataset is the graphQL query to obtain static dataset counts
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

// QueryDimensionOptions is the graphQL query to obtain static dataset dimension options
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

// GraphQL Query to obtain dimensions
const QueryDimensions = `
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
						categories {
							totalCount
						}
					}
				}
			}
		}
	}
}`
