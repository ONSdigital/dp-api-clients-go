package mock

const ErrorResponseNoDataset = `
{
	"data": {
	  "dataset": null
	},
	"errors": [
		{
			"message": "unknown dataset: \"test_dataset\" lang \"en\"",
			"locations": [
				{
					"line": 3,
					"column": 5
				}
			],
			"path": [
				"dataset"
			]
		}
	]
}
`

const GetDefaultClassicationResponseHappy = `
{
	"data": {
		"dataset": {
			"vars": [
		  		{
					"meta": {
			  			"Default_Classification_Flag": "N"
					},
					"name": "test_variable_1"
				},
				{
					"meta": {
			  			"Default_Classification_Flag": "Y"
					},
					"name": "test_variable_2"
				}
			]
		}
	}
}
`

const GetDefaultClassicationResponseMultipleDefaultVariables = `
{
	"data": {
		"dataset": {
			"vars": [
		  		{
					"meta": {
			  			"Default_Classification_Flag": "Y"
					},
					"name": "test_variable_1"
				},
				{
					"meta": {
			  			"Default_Classification_Flag": "Y"
					},
					"name": "test_variable_2"
				}
			]
		}
	}
}
`

const GetDefaultClassicationResponseNoDefaultVariables = `
{
	"data": {
		"dataset": {
			"vars": [
		  		{
					"meta": {
			  			"Default_Classification_Flag": "N"
					},
					"name": "test_variable_1"
				},
				{
					"meta": {
			  			"Default_Classification_Flag": "N"
					},
					"name": "test_variable_2"
				}
			]
		}
	}
}
`
