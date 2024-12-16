# Variables
PROJECT_ID = $(TF_VAR_project_id)
REGION = $(TF_VAR_region)
BUCKET_NAME = $(TF_VAR_bucket_name)-$(PROJECT_ID)
AGG_IMAGE_NAME = aggregator
TP_IMAGE_NAME = token-prices

# Binaries
AGG_BINARY = bin/aggregator
TP_BINARY = bin/token-prices

# Path
CLICKHOUSE_SCHEMA = schema/marketplace_analytics.sql
SAMPLE_DATA_PATH = sample_data/sample_data.csv

.PHONY: export-env deps lint build-local run-aggregator run-token-prices build push terraform-init terraform-apply deploy clean test apply-schema upload-test-data

# Load environment variables from .env
export-env:
	@echo "Loading environment variables from .env"
	@export $(shell cat .env | xargs)

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	@echo "Dependencies installed."

# Lint the code
lint:
	@echo "Running linters..."
	@golangci-lint run ./... --timeout 5m
	@echo "Linting complete."

# Build the aggregator and token-prices binaries locally
build-local: deps lint
	@echo "Building the aggregator binary..."
	@mkdir -p bin
	@go build -o $(AGG_BINARY) ./cmd/aggregator/main.go
	@echo "Aggregator binary built at $(AGG_BINARY)."
	@echo "Building the token-prices binary..."
	@go build -o $(TP_BINARY) ./cmd/token-prices/main.go
	@echo "Token-prices binary built at $(TP_BINARY)."

# Run the Aggregator pipeline locally
run-aggregator: build-local
	@echo "Running Aggregator pipeline locally..."
	@echo "Step 1: Extract"
	$(AGG_BINARY) extract \
		--input=$(EXTRACT_INPUT) \
		--output=$(EXTRACT_OUTPUT)
	@echo "Step 2: Transform"
	$(AGG_BINARY) transform \
		--input=$(EXTRACT_OUTPUT) \
		--output=$(TRANSFORM_OUTPUT) \
		--rates=$(TRANSFORM_RATES)
	@echo "Step 3: Load"
	$(AGG_BINARY) load \
		--input=$(TRANSFORM_OUTPUT) \
		--destination=$(LOAD_DESTINATION)
	@echo "Aggregator pipeline completed successfully."

# Run token-prices locally
run-token-prices-local: build-local
	@echo "Running token-prices locally..."
	@$(TP_BINARY) $(COMMAND) --tokens=$(TOKENS) --output=$(OUTPUT)

# Apply schema to ClickHouse
apply-schema: export-env
	@echo "Applying schema to ClickHouse..."
	clickhouse-client \
		--host=$(CLICKHOUSE_HOST) \
		--port=$(CLICKHOUSE_PORT) \
		--user=$(CLICKHOUSE_USER) \
		--password=$(CLICKHOUSE_PASSWORD) \
		--query="CREATE DATABASE IF NOT EXISTS analytics;"
	clickhouse-client \
		--host=$(CLICKHOUSE_HOST) \
		--port=$(CLICKHOUSE_PORT) \
		--user=$(CLICKHOUSE_USER) \
		--password=$(CLICKHOUSE_PASSWORD) \
		--database=analytics \
		--multiquery < $(CLICKHOUSE_SCHEMA)
	@echo "Schema applied successfully to ClickHouse."


# Upload test data to GCP bucket
upload-test-data: export-env
	@echo "Uploading test data to GCP bucket $(BUCKET_NAME)..."
	gsutil cp $(SAMPLE_DATA_PATH) gs://$(BUCKET_NAME)/
	@echo "Test data uploaded successfully."

# Build Docker images for both services
build: export-env
	@echo "Building Docker images..."
	docker build -t gcr.io/$(PROJECT_ID)/$(AGG_IMAGE_NAME):latest -f Dockerfile.aggregator .
	docker build -t gcr.io/$(PROJECT_ID)/$(TP_IMAGE_NAME):latest -f Dockerfile.token-prices .
	@echo "Docker images built."

# Push Docker images for both services
push: export-env
	@echo "Pushing Docker images to Google Container Registry..."
	docker push gcr.io/$(PROJECT_ID)/$(AGG_IMAGE_NAME):latest
	docker push gcr.io/$(PROJECT_ID)/$(TP_IMAGE_NAME):latest
	@echo "Docker images pushed to GCR."

# Terraform initialization
terraform-init: export-env
	@echo "Initializing Terraform..."
	terraform init

# Terraform apply
terraform-apply: export-env
	@echo "Applying Terraform configuration..."
	terraform apply -auto-approve

# Full deployment (build, push, Terraform)
deploy: build push terraform-init terraform-apply
	@echo "Deployment complete."

# Clean local environment
clean:
	@echo "Cleaning up local environment..."
	@rm -rf bin
	@docker rmi -f gcr.io/$(PROJECT_ID)/$(AGG_IMAGE_NAME):latest || true
	@docker rmi -f gcr.io/$(PROJECT_ID)/$(TP_IMAGE_NAME):latest || true
	@rm -rf .terraform .terraform.lock.hcl terraform.tfstate terraform.tfstate.backup || true
	@echo "Cleanup complete."

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v -race -count=1
	@echo "All tests passed."