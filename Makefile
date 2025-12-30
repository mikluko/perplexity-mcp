.DEFAULT_GOAL := help

.PHONY: help build build-image build-chart test test-coverage lint clean install run-stdio run-http

# Variables
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-$(shell git rev-parse --short HEAD)")
VERSION := $(GIT_VERSION:v%=%)
KO_DOCKER_REPO := ghcr.io/mikluko/perplexity-mcp
CHART_REPO := ghcr.io/mikluko/helm-charts
export KO_DOCKER_REPO

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: image package ## Build and push container image and Helm chart

image: ## Build and push container image with ko
	@echo "Building and pushing image: $(KO_DOCKER_REPO):$(VERSION)"
	VERSION=$(VERSION) ko build --bare --tags $(VERSION)

package: ## Package and push Helm chart to OCI registry
	@echo "Packaging and pushing chart: $(CHART_REPO)/perplexity-mcp:$(VERSION)"
	@helm package charts/perplexity-mcp --version $(VERSION) --app-version $(VERSION) --destination .build/
	@helm push .build/perplexity-mcp-$(VERSION).tgz oci://$(CHART_REPO)
	@rm .build/perplexity-mcp-$(VERSION).tgz
	@echo "Chart pushed successfully"

# Test targets
test: ## Run tests
	go test ./... -v

test-coverage: ## Run tests with coverage
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	golangci-lint run ./...

# Build targets
clean: ## Clean build artifacts
	rm -f perplexity-mcp coverage.out coverage.html
	rm -rf dist/

install: ## Install the binary
	go install -ldflags="-X main.version=$(VERSION)"

# Local development
run-stdio: ## Run in stdio mode locally
	@if [ -z "$(PERPLEXITY_API_KEY)" ]; then \
		echo "Error: PERPLEXITY_API_KEY not set"; \
		exit 1; \
	fi
	./perplexity-mcp -mode stdio

run-http: ## Run in HTTP mode locally
	@if [ -z "$(PERPLEXITY_API_KEY)" ]; then \
		echo "Error: PERPLEXITY_API_KEY not set"; \
		exit 1; \
	fi
	./perplexity-mcp -mode http -listen localhost:8080
