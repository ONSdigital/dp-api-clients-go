# Static File Upload Client

## Usage
### Setup

Local:
```go
c := files.NewAPIClient("http://localhost:26900")
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