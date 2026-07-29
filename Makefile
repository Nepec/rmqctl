BINARY_NAME := rmqctl
PKG := ./...

GOOS ?= linux
GOARCH ?= amd64
VERSION ?= $(shell git describe --tags --always --dirty)

EXT := $(if $(filter windows,$(GOOS)),.exe,)

# build rmqctl binary in the bin directory with the version set to the git tag
# or commit hash if there is no tag.
.PHONY: build 
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
			-ldflags "-X main.version=$(VERSION)" \
			-o bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(EXT) \
			./cmd/rmqctl

# run unit tests across all packages
.PHONY: test
test:
	go test -v -race -count=1 $(PKG)

# run integration tests using docker
.PHONY: integration docker-cleanup
RABBIT_CONTAINER := rabbitmq-integration
RMQ_MNG_PORT := 15672

integration:
	$(MAKE) docker-cleanup	# rm any orphaned containers from previously failed runs
	@ docker run -d --name $(RABBIT_CONTAINER) -p $(RMQ_MNG_PORT):15672 rabbitmq:management 1> /dev/null
	@ until curl -sf -u guest:guest "http://localhost:15672/api/overview" >/dev/null 2>&1; do \
		echo "Waiting for RabbitMQ management API to become ready..."; \
		sleep 2; \
	done
	@ go test -tags=integration -v ./integration/... ; \
		status=$$? ; \
		$(MAKE) docker-cleanup ; \
		exit $$status


docker-cleanup: # cleanup the testing container
	docker rm -f $(RABBIT_CONTAINER) >/dev/null 2>&1 || true

.PHONY: golangci-lint
golangci-lint:
	golangci-lint run --timeout 60m ./...

.PHONY: fmt
fmt:
	gofumpt -w .

.PHONY: clean
clean:
	rm -rf ./bin/
