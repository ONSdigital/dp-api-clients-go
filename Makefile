makSHELL=bash

test:
	go test -count=1 -race -cover ./...

.PHONY: test

audit: ## Runs checks for security vulnerabilities on dependencies (including transient ones)
	dis-vulncheck
.PHONY: audit

build:
	go build ./...
.PHONY: build

.PHONY: lint
lint:
	golangci-lint run ./...


