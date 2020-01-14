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
* importapi
* renderer
* search
* zebedee

### Client healthchecks

Each client has it's own healthcheck function; this will soon be deprecated as applications/services will use the new Checker functions which implement the health package.

List of clients:

* codelist
* dataset
* filter
* hierarchy
* importapi
* search
* zebedee

If a service does not have a client library use the abstracted Checker function in health package like so:

```
import "github.com/ONSdigital/dp-api-clients-go/health"

func main() {
    ...
    // Create new healthcheck rchttp client, this will set the `/health` and `/healthcheck` as endpoints that are not retiable
    hcClient := health.NewClient(<name>, <url>)

    ctx := context.Background()

    // Check state of external service
    checkObj, err := hcClient.Checker(ctx)
    if err != nil {
        ...
    }
    
    ...
}

```

### Tests

Run tests using `make test`

### Licence

Copyright ©‎ 2019, Crown Copyright (Office for National Statistics) (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.
