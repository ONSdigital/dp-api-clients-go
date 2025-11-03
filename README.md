# dp-api-clients-go [![GoDoc](https://godoc.org/github.com/ONSdigital/dp-api-clients-go/v2?status.svg)](https://godoc.org/github.com/ONSdigital/dp-api-clients-go/v2)

Common client code - in go - for ONS APIs:

* areas
* clientlog - logging
* codelist
* dataset
* filter
* headers - common API request headers
* healthcheck -> health
* hierarchy
* identity
* image
* importapi
* releasecalendar
* renderer
* search (dimension search)
* site-search (deprecated in favour of [dp-search-api SDK](https://github.com/ONSdigital/dp-search-api/tree/develop/sdk))
* upload (Static Files)

## Usage

Each client defines two constructor functions: one that creates a new healthcheck client (with a new dp-net/http Clienter), and the other that allows you to provide it externally, so that you can reuse it among different clients.

For example, you may create a new image API client like so:

```go
    import  "github.com/ONSdigital/dp-api-clients-go/v2/image"

    ...
    imageClient := image.NewAPIClient(<url>)
    ...
```

Or you may create it providing a Healthcheck client:

```go
    import  "github.com/ONSdigital/dp-api-clients-go/v2/image"
    import  "github.com/ONSdigital/dp-api-clients-go/v2/health"

    ...
    hcClient := health.NewClient(<genericName>, <url>)
    imageClient := image.NewWithHealthClient(hcClient)
    ...
```

### Batch processing

Each method in each client corresponds to a single call against one endpoint of an API, except for the Batch processing calls, which may trigger multiple concurrent calls.

The batch processing logic is implemented in the batch package as a generic method (`ProcessInConcurrentBatches`) that can be used by multiple client implementations to handle the processing of paginated responses.

For each batch, a parallel go-routine will trigger the provided getter method (`GenericBatchGetter`). Once the getter method returns, the resulting batch is provided to the processor method (`GenericBatchProcessor`) after acquiring a lock to guarantee mutually exclusive execution of processors.

The algorithm can be configured with a maximum number of items per batch (which will control the offset of each getter call) and a maximum number of workers, which will limit the number of concurrent go-routines that are executed at the same time.

If any getter or processor returns an error, the algorithm will be aborted and the same error will be returned. The processor may also return a boolean value of `true` to force the abortion of the algorithm, even if there is no error.

So far, the batch processing has been implemented by `filter API` and `dataset API` clients in order to obtain dimension options.

#### Get in batches

Assuming you have a dataset client called `datasetClient`, then you can get all the options in batches like so:

```go
    // obtain all options after aggregating paginated GetOption responses
    allValues, err := datasetClient.GetOptionsInBatches(ctx, userToken, serviceToken, collectionID, datasetID, edition, version, dimensionName, batchSize, maxWorkers)
```

where `batchSize` is the maximum number of items requested in each batch, and `maxWorkers` is the maximum number of concurrent go-routines.
This method will call `GET options` for each batch and then it will aggregate the results until we have all the options.

Instead of aggregating the results, you may want to perform some different logic for each batch. In this case, you may use `GetOptionsBatchProcess` with your batch Processor, like so:

```go
    
    // processBatch is a function that performs some logic for each batch, and has the ability to abort execution if forceAbort is true or an error is returned.
    var processBatch dataset.OptionsBatchProcessor = func(batch dataset.Options) (forceAbort bool, err error) {
        // <Do something with batch>
        return false, nil
    }

    // list of option IDs to obtain (if nil, all options will be provided)
    optionIDs := []string{"option1", "option2", "option3"}
    
    // call dataset API GetOptionsBatchProcess with the batch processor
    err = f.DatasetClient.GetOptionsBatchProcess(ctx, userToken, serviceToken, collectionID, datasetID, edition, version, dimensionName, &optionIDs, processBatch, f.maxDatasetOptions, f.BatchMaxWorkers)
    return idLabelMap, err
```

## Package docs

* [health](health/README.md#health)

## Tests

Run tests using `make test`

## Licence

Copyright ©‎ 2021, Crown Copyright (Office for National Statistics) <https://www.ons.gov.uk>

Released under MIT license, see [LICENSE](LICENSE.md) for details.
