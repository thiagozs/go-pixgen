SHELL := /bin/bash

APP_NAME := pixgen
CLI_MAIN := ./cmd/pixgen
CLI_BINARY := bin/$(APP_NAME)
DOCKER_IMAGE := $(APP_NAME):latest
GO_FILES := $(shell find . -name '*.go' -not -path "./.gocache/*" -not -path "./internal/cobra/testdata/*")
ARGS ?=

.PHONY: help build test fmt cli run docker-build docker-run clean

help:
	@echo "Available targets:"
	@echo "  make build         - Compile all Go packages"
	@echo "  make test          - Run unit tests"
	@echo "  make fmt           - Format Go code with gofmt"
	@echo "  make cli           - Build CLI binary into $(CLI_BINARY)"
	@echo "  make run ARGS='...' - Run CLI with optional arguments (e.g., ARGS=\"serve --addr :8080\")"
	@echo "  make docker-build  - Build CLI Docker image ($(DOCKER_IMAGE))"
	@echo "  make docker-run    - Run CLI Docker image"
	@echo "  make clean         - Remove build artifacts and Docker image"

build:
	@echo "==> Building Go packages"
	go build ./...

test:
	@echo "==> Running tests"
	GOCACHE=$$(pwd)/.gocache go test ./...

fmt:
	@echo "==> Formatting code"
	gofmt -w $(GO_FILES)

cli: $(CLI_BINARY)

$(CLI_BINARY): $(GO_FILES)
	@echo "==> Building CLI binary $@"
	@mkdir -p $(@D)
	CGO_ENABLED=0 go build -o $@ $(CLI_MAIN)

run:
	@echo "==> Running CLI $(ARGS)"
	go run $(CLI_MAIN) $(ARGS)

docker-build:
	@echo "==> Building Docker image $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	@echo "==> Running Docker image $(DOCKER_IMAGE)"
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

clean:
	@echo "==> Cleaning artifacts"
	rm -rf $(CLI_BINARY) .gocache
	-@docker rmi $(DOCKER_IMAGE) >/dev/null 2>&1 || true
