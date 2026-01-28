MODULE_NAME := fcstask

.PHONY: init tidy gen test migrate docker-up docker-down docker-logs clean

init:
	@if [ ! -f go.mod ]; then \
		echo "Init repo: $(MODULE_NAME)"; \
		go mod init $(MODULE_NAME); \
	else \
		echo "good. already exists"; \
	fi

tidy:
	go mod tidy

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