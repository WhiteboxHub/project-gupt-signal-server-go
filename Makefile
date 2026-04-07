.PHONY: help build run test clean docker-build docker-run deploy-gcp

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	go build -o bin/signaling-server ./cmd/server

run: ## Run the server locally
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

# Test clients
run-host: ## Run host test client
	go run ./cmd/testclient/host.go

run-client: ## Run client test client
	go run ./cmd/testclient/client.go

# Docker commands
docker-build: ## Build Docker image
	docker build -t signaling-server:latest .

docker-run: ## Run Docker container locally
	docker run -p 8080:8080 signaling-server:latest

# Google Cloud deployment
GCP_PROJECT ?= my-project
GCP_REGION ?= us-central1
IMAGE_NAME = gcr.io/$(GCP_PROJECT)/signaling-server

deploy-gcp: ## Deploy to Google Cloud Run
	@echo "Building and deploying to GCP..."
	gcloud builds submit --tag $(IMAGE_NAME)
	gcloud run deploy signaling-server \
		--image $(IMAGE_NAME) \
		--platform managed \
		--region $(GCP_REGION) \
		--allow-unauthenticated \
		--port 8080

# Dependencies
deps: ## Download dependencies
	go mod download
	go mod tidy

# Linting
lint: ## Run linter
	golangci-lint run
