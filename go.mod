module github.com/ONSdigital/dp-api-clients-go/v2

go 1.18

require (
	github.com/ONSdigital/dp-healthcheck v1.5.0
	github.com/ONSdigital/dp-mocking v0.9.1
	github.com/ONSdigital/dp-net v1.5.0
	github.com/ONSdigital/dp-net/v2 v2.8.0
	github.com/ONSdigital/log.go/v2 v2.3.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/graphql v0.0.0-20220606043923-3cf50f8a0a29
	github.com/smartystreets/goconvey v1.7.2
)

require (
	github.com/ONSdigital/dp-api-clients-go v1.43.0 // indirect
	github.com/ONSdigital/log.go v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.44.204 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

retract (
	[v2.226.0,v2.227.0] // contains breaking code
)
