# Natural Language Processing microservices 

## Description 
Incorporation of Natural Language Processing (NLP) into dp-search-api as a search engine. It's powered by three microservices: dp-berlin-api, dp-category-api, and dp-scrubber-api. 
These services work together to upgrade the search experience. By utilizing NLP, the system delivers more accurate and intuitive search results, making searches more effective and user-friendly.

Our aim is to optimize search results without compromising the current functionality.

Berlin and Category are python microservices which is why their Go clients have been added to dp-api-clients-go with more description below. 

## Berlin Client

### Description

Responsible for identifying geospational data for more information on the api read [this README.md](https://github.com/ONSdigital/dp-nlp-berlin-api/blob/develop/README.md)

### Usage

[Check this readme](berlin/README.md)

## Category Client

### Description

Responsible for categorisation of the query for more information on the api read [this README.md](https://github.com/ONSdigital/dp-nlp-category-api#readme)

### Usage

[Check this readme](category/README.md)


## Scrubber Client

### Description

This API allows users to identify Output Areas (OA) and Industry Classification (SIC) associated with a given location. OAs are small geographical areas in the UK used for statistical purposes, while SIC codes are a system of numerical codes used to identify and categorize industries. 


Scrubber is written in go so has it's own Go SDK available [here.](https://github.com/ONSdigital/dp-search-scrubber-api#readme)

### Usage

[Check this readme](category/README.md)