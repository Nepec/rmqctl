BINARY_NAME := rmqctl
PKG := ./...

GOOS ?= linux
GOARCH ?= amd64
VERSION ?= $(shell git describe --tags --always --dirty)

EXT := $(if $(filter windows,$(GOOS)),.exe,)

.PHONY: build test lint fmt clean tidy
 
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			 -ldflags "-X rmqctl/version.Version=$(VERSION)" \
			-o bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(EXT) \
			./cmd/rmqctl

test:
	go test -v -race -count=1 $(PKG)

lint:
	golangci-lint run --timeout 60m ./...

fmt:
	gofumpt -w .

tidy:
	go mod tidy
 
clean:
	rm -rf ./bin/
