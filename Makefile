GOPATH := $(shell go env GOPATH)
PATH := $(PATH):$(GOPATH)/bin
MODULE_NAME := fcstask-backend
BINARY_NAME := fcstask-api
DOCKER_IMAGE_NAME ?= miruken/$(MODULE_NAME)-backend
DOCKER_IMAGE_TAG ?= 0.1.0

.PHONY: init tidy migrate migrate install-tools gen test test-integration-db postgreplication-up docker-build docker-run docker-test docker-push ci-local ci

init:
	@echo "🔧 Initializing repo: $(MODULE_NAME)..."
	@if [ ! -f go.mod ]; then \
		go mod init $(MODULE_NAME) && \
		echo "✅ go.mod created"; \
	else \
		echo "⚠️  go.mod already exists"; \
	fi

tidy:
	@echo "🧹 Tidying dependencies..."
	@go mod tidy
	@echo "✅ go.mod & go.sum updated"

# Миграции БД
migrate:
	@echo "Running database migrations..."
	go run ./cmd/migrate/main.go

install-tools:
	@echo "📦 Installing tools..."
	@which oapi-codegen || go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@which mockgen || go install github.com/golang/mock/mockgen@latest
	@go get github.com/golang/mock/gomock
	@echo "✅ Tools installed"

gen: install-tools
	@echo "Generating API code from OpenAPI..."
	@if command -v oapi-codegen >/dev/null 2>&1; then \
		echo "oapi-codegen is already installed"; \
	else \
		echo "Installing oapi-codegen..."; \
		go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest; \
	fi
	@echo "Generating types..."
	oapi-codegen -generate types -package api -o internal/api/types.gen.go api/openapi.yaml
	@echo "Generating server..."
	oapi-codegen -generate server -package api -o internal/api/server.gen.go api/openapi.yaml
	@echo "Code generation completed!"
	@echo "🔄 Generating code..."
	@go generate ./...
	@echo "✅ Code generation completed"

test: gen
	@echo "🧪 Running tests..."
	@go test ./... -v
	@echo "✅ Tests completed"

postgreplication-up:
	podman compose -f compose.postgres-replication.yaml up -d

test-integration-db:
	go test -tags=integration -count=1 -v ./internal/db/...

docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .
	@echo "✅ Docker image built"

docker-run: docker-build
	@echo "🚀 Running container on http://localhost:8080"
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

docker-test:
	@echo "🧪 Running tests inside container..."
	docker run --rm \
		-v "$(PWD):/app" \
		-w /app \
		golang:1.25-alpine \
		go test ./... -v

docker-push:
	@if [ -z "$$CI" ] && [ -z "$$FORCE_PUSH" ]; then \
		echo "🛑 ERROR: Refusing to push from local machine."; \
		echo "💡 Run with FORCE_PUSH=1 to override (not recommended)."; \
		exit 1; \
	fi
	@echo "📤 Pushing image to registry..."
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@echo "✅ Pushed: $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)"

ci-local: init tidy test docker-build
	@echo "✅ Local CI check passed!"

ci: ci-local docker-push
	@echo "✅ Full CI pipeline completed!"
