## Berlin Client

### Description

Responsible for geospational data for more information read [this README.md](https://github.com/ONSdigital/dp-nlp-berlin-api/blob/develop/README.md)

### Usage

#### Initialization

```go
import "github.com/ONSdigital/dp-api-clients-go/v2/nlp/berlin"

// Initialize the Berlin API client
client := berlin.New("https://berlinURL.com")
```

#### Functionality

signiture
```go
// GetBerlin gets a list of berlin results based on the berlin request
// you can get options from the berlin client package 
func (cli *Client) GetBerlin(ctx context.Context, options Options) (*models.Berlin, errors.Error)
```
Once initialized you can make a request to Berlin like so:

```go
// Create an Options struct and set a query parameter 'q'
// you can also use url.Values directly into the Options
options := berlin.Options{}
options.Q("your_query_here")

// Add custom headers to the options
options.Headers = http.Header{}
options.Headers.Set(auithHeader, "")
options.Headers.Set(someOtherHeader, "")

// Get Berlin results using the created client and custom options
results, err := client.GetBerlin(ctx, options)
if err != nil {
    // handle error
}
```

you can reuse a healthcheck client like so:

```go
berlinClient := client.NewWithHealthClient(hcCli)
// same thing next
```

some other functionality:
```go
// URL returns the URL used by this client
func (cli *Client) URL() string {

// Health returns the underlying Healthcheck Client for this berlin API client
func (cli *Client) Health() *healthcheck.Client {

// Checker calls berlin api health endpoint and returns a check object to the caller
func (cli *Client) Checker(ctx context.Context, check *health.CheckState) error {
```
