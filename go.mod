module github.com/ONSdigital/dp-api-clients-go/v2

go 1.18

require (
	github.com/ONSdigital/dp-healthcheck v1.6.0
	github.com/ONSdigital/dp-mocking v0.10.0
	github.com/ONSdigital/dp-net/v2 v2.9.0
	github.com/ONSdigital/log.go/v2 v2.4.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/graphql v0.0.0-20220606043923-3cf50f8a0a29
	github.com/smartystreets/goconvey v1.8.0
)

require (
	github.com/fatih/color v1.15.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/smartystreets/assertions v1.13.1 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

retract [v2.226.0, v2.227.0] // contains breaking code
