## Category Client

### Description

Responsible for categorisation of the query for more information on the api [this README.md](https://github.com/ONSdigital/dp-nlp-category-api#readme)

### Usage

#### Initialization

```go
import "github.com/ONSdigital/dp-api-clients-go/v2/nlp/category"

// Initialize the Category API client
client := category.New("https://categoryURL.com")
```

#### Functionality

signature
```go
// GetCategory gets a list of category results based on the category request
// you can get options from the category client package 
func (cli *Client) GetCategory(ctx context.Context, options Options) (*models.Category, errors.Error)
```
Once initialized you can make a request to Category like so:

```go
// Create an Options struct and set a query parameter 'query'
// you can also use url.Values directly into the Options
options := category.OptInit()
options.Q("your_query_here")

// Add custom headers to the options
options.Headers = http.Header{}
options.Headers.Set(authHeader, "")
options.Headers.Set(someOtherHeader, "")

// Get Category results using the created client and custom options
results, err := client.GetCategory(ctx, options)
if err != nil {
    // handle error
}
```

you can reuse a healthcheck client like so:

```go
categoryClient := client.NewWithHealthClient(hcCli)
// same thing next
```

some other functionality:
```go
// URL returns the URL used by this client
func (cli *Client) URL() string {

// Health returns the underlying Healthcheck Client for this category API client
func (cli *Client) Health() *healthcheck.Client {

// Checker calls category api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
```
