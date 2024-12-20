.PHONY: build run clean test dev logs watch

# Build variables
BINARY_NAME=news-reader
BUILD_DIR=build

# Version information
VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-w -s \
	-X github.com/news-reader/internal/version.Version=$(VERSION) \
	-X github.com/news-reader/internal/version.BuildTime=$(BUILD_TIME) \
	-X github.com/news-reader/internal/version.GitCommit=$(GIT_COMMIT)"

all: clean build

build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v ./...

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

# Development targets
dev:
	./scripts/dev.sh

logs:
	./scripts/logs.sh

# Docker targets
docker-build:
	podman build -t $(BINARY_NAME) .

docker-run:
	podman run -d \
		--name $(BINARY_NAME) \
		-p 8082:8082 \
		-v news-reader-data:/data \
		$(BINARY_NAME)

docker-stop:
	podman stop $(BINARY_NAME) || true
	podman rm $(BINARY_NAME) || true

docker-logs:
	podman logs -f $(BINARY_NAME)
