dp-api-clients-go [![Build Status](https://travis-ci.org/ONSdigital/dp-api-clients-go.svg?branch=master)](https://travis-ci.org/ONSdigital/dp-api-clients-go) [![GoDoc](https://godoc.org/github.com/ONSdigital/dp-api-clients-go?status.svg)](https://godoc.org/github.com/ONSdigital/dp-api-clients-go)
=====

Common client code - in go - for ONS APIs:

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
* renderer
* search


#### Usage

(WIP) Each client defines two constructor functions: one that creates a new dp-net/http Clienter, and the other that allows you to provide it externally, so that you can reuse it among different clients.

For example, you may initialise an image API client with a new Clienter:
```
    import  "github.com/ONSdigital/dp-api-clients-go/image"
    ...
    imageClient := image.NewAPIClient(<url>)
    ...
```

Or you may initialise it providing a Clienter:
```
    ...
    imageClient := image.NewWithClienter(<url>, <clienter>)
    ...
```

#### Package docs

* [health](https://github.com/ONSdigital/dp-api-clients-go/tree/feature/client-checker/health)

### Tests

Run tests using `make test`

### Licence

Copyright ©‎ 2019, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
