module github.com/ONSdigital/dp-api-clients-go/v2

go 1.23

require (
	github.com/ONSdigital/dp-healthcheck v1.6.3
	github.com/ONSdigital/dp-mocking v0.10.1
	github.com/ONSdigital/dp-net/v2 v2.11.2
	github.com/ONSdigital/log.go/v2 v2.4.3
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/graphql v0.0.0-20230722043721-ed46e5a46466
	github.com/smartystreets/goconvey v1.8.1
)

require (
	github.com/fatih/color v1.17.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

retract [v2.226.0, v2.227.0] // contains breaking code
