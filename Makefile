.DEFAULT_GOAL := help

GO ?= go
GOLANGCI_LINT ?= golangci-lint
DOCKER ?= docker

BIN ?= dbot
BIN_DIR ?= bin

CACHE_DIR ?= .cache
GO_BUILD_CACHE ?= $(CURDIR)/$(CACHE_DIR)/go-build
GOLANGCI_LINT_CACHE ?= $(CURDIR)/$(CACHE_DIR)/golangci-lint

IMAGE ?= dbot:local

.PHONY: help
help:
	@printf "%s\n" \
		"Targets:" \
		"  make fmt           Format Go code (gofmt)" \
		"  make test          Run unit tests" \
		"  make lint          Run golangci-lint" \
		"  make build         Build binary to $(BIN_DIR)/$(BIN)" \
		"  make run           Run (go run .)" \
		"  make docker-build  Build Docker image ($(IMAGE))" \
		"  make docker-test   Run tests in Docker build stage" \
		"  make docker-run    Run Docker image (requires DISCORD_TOKEN)" \
		"  make clean         Remove build artifacts and caches"

.PHONY: fmt
fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

.PHONY: test
test:
	mkdir -p $(CACHE_DIR)/go-build
	GOCACHE="$(GO_BUILD_CACHE)" $(GO) test ./...

.PHONY: lint
lint:
	mkdir -p $(CACHE_DIR)/golangci-lint $(CACHE_DIR)/go-build
	GOCACHE="$(GO_BUILD_CACHE)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" $(GOLANGCI_LINT) run ./...

.PHONY: build
build:
	mkdir -p $(BIN_DIR) $(CACHE_DIR)/go-build
	GOCACHE="$(GO_BUILD_CACHE)" CGO_ENABLED=0 $(GO) build -trimpath -o "$(BIN_DIR)/$(BIN)" .

.PHONY: run
run:
	$(GO) run .

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 $(DOCKER) build -t "$(IMAGE)" .

.PHONY: docker-test
docker-test:
	DOCKER_BUILDKIT=1 $(DOCKER) build --target test .

.PHONY: docker-run
docker-run:
	$(DOCKER) run --rm -e DISCORD_TOKEN "$(IMAGE)"

.PHONY: clean
clean:
	rm -rf "$(BIN_DIR)" "$(CACHE_DIR)"
