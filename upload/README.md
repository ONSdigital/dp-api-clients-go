# Static File Upload Client

## Usage
### Setup

Local:
```go
c := upload.NewAPIClient("http://localhost:25100")
```

Remote: 
```go
c := upload.NewAPIClient("http://localhost:11850")
```

### Uploading a file

```go
f := io.NopCloser(strings.NewReader("File content"))
m := upload.Metadata{
    CollectionID:  &collectionID, // Collection ID is option. Leave it unset if you do not have it at upload
    FileName:      "test.txt",
    Path:          "testing/docs",
    IsPublishable: true,
    Title:         "A testing file",
    FileSizeBytes: 12,
    FileType:      "text/plain",
    License:       "MIT",
    LicenseURL:    "https://opensource.org/licenses/MIT",
}

err := c.Upload(f, m)

if err != nil {
	...
}
```