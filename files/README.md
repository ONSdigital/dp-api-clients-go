# Static File Upload Client

## Usage
### Setup

Local:
```go
c := files.NewAPIClient("http://localhost:26900")
```

If using a healthcheck client, it can be instantiated as follows:

```go
import "github.com/ONSdigital/dp-api-clients-go/v2/health"

hcClient := health.NewClient("api-router", "http://localhost:26900")
c := files.NewWithHealthClient(hcClient)
```

Remote:
```go
c := files.NewAPIClient("http://localhost:12700")
```

### Set collection ID

```go
err := c.SetCollectionID(context.Background(), "testing/test.txt", "123456789")

if err != nil {
	...
}
```

### Publish Collection

```go
err := c.PublishCollection(context.Background(), "123456789")

if err != nil {
    ...
}
```

### Get File by Path

```go
result, err := client.GetFile(context.Background(), "test/testing.csv", "AUTH TOKEN")

if err != nil {
    ...
}

if result.IsPublshable {
	...
}
```

