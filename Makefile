SHELL := /bin/bash

# ============================================================================
# HELPERS
# ============================================================================

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:N} = y ]

# ============================================================================
# CI 
# ============================================================================

## vendor: tidy and vendor deps
.PHONY: tidy
tidy:
	@echo 'Tidying and verifying deps...'
	go mod tidy
	go mod verify
	@echo 'Vendoring deps...'
	go mod vendor

## ci: tidy, vendor, fmt, vet and test
.PHONY: ci
ci: tidy
	@echo Formatting code...
	go fmt ./...
	@echo Vetting code...
	go vet ./...
	golangci-lint run
	@echo Running tests...
	go test -race -vet=off ./...

# ============================================================================
# TEST
# ============================================================================
## cover: build coverage profile at p.out with go test
cover:
	go test -coverprofile p.out

## cover-show: open webpage showing coverage 
cover-show: cover
	go tool cover -html p.out

