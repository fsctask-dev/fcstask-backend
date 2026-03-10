GOPATH := $(shell go env GOPATH)
PATH := $(PATH):$(GOPATH)/bin
MODULE_NAME := fcstask-backend
BINARY_NAME := fcstask-api
DOCKER_IMAGE_NAME ?= miruken/$(MODULE_NAME)-backend
DOCKER_IMAGE_TAG ?= 0.1.0

.PHONY: init tidy gen test migrate docker-up docker-down docker-logs clean

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

# Генерация API кода
gen:
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

test:
	go test ./... -v

# Миграции БД
migrate:
	@echo "Running database migrations..."
	go run ./cmd/migrate/main.go

run:
	go run ./internal/cmd/main.go

build:
	go build -o bin/server ./internal/cmd/main.go

# Docker команды
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Очистка сгенерированных файлов
clean:
	rm -f internal/api/*.gen.go
