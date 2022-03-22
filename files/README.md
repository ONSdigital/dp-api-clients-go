# Static File Upload Client

## Usage
### Setup

```go
c := files.NewAPIClient("http://localhost:26900")
```

### Set collection ID

```go
err := c.SetCollectionID(context.Background(), "testing/test.txt", "123456789")

if err != nil {
	...
}
```