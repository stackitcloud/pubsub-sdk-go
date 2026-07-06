GOCMD=go
GO_VERSION := $(shell $(GOCMD) version | awk '{print $$3}' | sed 's/go//')
SHELL := /bin/bash
.SHELLFLAGS += -o pipefail

export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: generate
generate:
	@echo "Generating code..."
	oapi-codegen --config=scripts/sdk.cfg.yaml api/openapi.yaml

.PHONY: prepare
prepare:
	go mod tidy
	golangci-lint run --fix

.PHONY: test
test: prepare
	rm -rf tmp/coverage
	mkdir -p tmp/coverage
	go test --race -coverpkg=./... -cover ./... -args -test.gocoverdir=$(CURDIR)/tmp/coverage
	@echo
	@echo "========== Correct coverage over all packages =========="
	go tool covdata percent -i=tmp/coverage
	go tool covdata textfmt -i=tmp/coverage -o tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html