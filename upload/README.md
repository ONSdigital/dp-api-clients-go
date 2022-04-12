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

Notes:
- all fields are required, except collection Id
  - but it must be set before publishing 
  - see https://github.com/ONSdigital/dp-api-clients-go/tree/main/files
- File name will be part of the AWS S3 object name so should adhere to the S3 object naming guidelines 
  - see https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-keys.html#:~:text=and%20AWS%20SDKs.-,Object%20key%20naming%20guidelines,-You%20can%20use
- Path will be part of the AWS S3 bucket name so should adhere to the S3 bucket naming rules 
  - see https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html
- FileType is the the mime type of the file being uploaded 
  - used when file is downloaded