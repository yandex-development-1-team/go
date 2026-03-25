# The name of the binary file
BIN_NAME=app

# Go params
GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GORUN=$(GO) run

# Paths
SRC_DIR=./cmd/bot
BUILD_DIR=./bin
GO_FILES=$(shell find . -type f -name '*.go' ! -path './vendor/*')

# Docker Compose (локальный стек: loopback + healthcheck в overlay)
COMPOSE_FILE=docker-compose.yml
COMPOSE_INFRA=docker-compose.infra.yml
COMPOSE_LOCAL_INFRA=docker-compose.local.infra.yml
COMPOSE_LOCAL_APP=docker-compose.local.app.yml

# Фронтенд в отдельном репозитории: make run-frontend FRONTEND_DIR=../your-frontend
FRONTEND_DIR?=
FRONTEND_SCRIPT_FULL?=dev
FRONTEND_SCRIPT_LOCAL_API?=dev:local-api

.PHONY: migration migration-create migration-rollback generate-mocks fmt fix lint lint-fix vet help \
	run-local run-local-api stop-local run-frontend run-frontend-local-api

all: build

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'

## build: building a project
build:
	$(GOBUILD) -C $(SRC_DIR) -o $(BUILD_DIR)/$(BIN_NAME)

## run: run the application
run:
	$(GORUN) $(SRC_DIR)/*.go

## run-local: Docker — инфра + app (бот, если API_ONLY=false). Порты на 127.0.0.1
run-local:
	@test -f .env || (echo "Создайте .env из .env.example"; exit 1)
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_LOCAL_INFRA) -f $(COMPOSE_LOCAL_APP) up -d --build

## run-local-api: Docker — только Postgres и Redis на 127.0.0.1 (бэкенд: make run + env из .env.local.api)
run-local-api:
	@test -f .env || (echo "Создайте .env из .env.example"; exit 1)
	docker compose -f $(COMPOSE_INFRA) -f $(COMPOSE_LOCAL_INFRA) up -d

## stop-local: остановить контейнеры локального проекта compose
stop-local:
	docker compose -f $(COMPOSE_FILE) -f $(COMPOSE_LOCAL_INFRA) -f $(COMPOSE_LOCAL_APP) down

## run-frontend: npm в FRONTEND_DIR (полный бэкенд в Docker после run-local)
run-frontend:
	@test -n "$(FRONTEND_DIR)" || (echo "Задайте FRONTEND_DIR=путь/к/фронту"; exit 1)
	cd "$(FRONTEND_DIR)" && npm run $(FRONTEND_SCRIPT_FULL)

## run-frontend-local-api: npm в FRONTEND_DIR (инфра в Docker, API на хосте)
run-frontend-local-api:
	@test -n "$(FRONTEND_DIR)" || (echo "Задайте FRONTEND_DIR=путь/к/фронту"; exit 1)
	cd "$(FRONTEND_DIR)" && npm run $(FRONTEND_SCRIPT_LOCAL_API)

## clean: delete binary
clean:	
	$(GOCLEAN)
	rm -rf $(SRC_DIR)/bin
	
## test: running tests
test:
	$(GOTEST) ./... -v -cover -count=1

## fmt: Format code and fix imports (goimports + gofmt)
fmt: fix

## fix: Fix imports (grouping, unused) and format code. Run before commit.
fix:
	$(GO) run golang.org/x/tools/cmd/goimports@latest -w $(GO_FILES)
	$(GO) fmt ./...

## vet: Check the code for suspicious structures
vet:
	$(GO) vet ./...

## lint: Run linter (golangci-lint). Install: https://golangci-lint.run/usage/install/
lint:
	golangci-lint run ./...

## lint-fix: Run linter with auto-fix where possible
lint-fix:
	golangci-lint run --fix ./...


migration:
	@echo "Migration commands:"
	@echo "  make migration-create NAME=<name>    Create new migration file"
	@echo "  make migration-rollback DB_DSN=<dsn> Rollback last migration"
	@echo ""
	@echo "Examples:"
	@echo "  make migration-create NAME=create_users_table"
	@echo "  make migration-rollback DB_DSN=<dsn>"

migration-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: migration name is required"; \
		echo "Usage: make migration-create NAME=migration_name"; \
		exit 1; \
	fi
	@mkdir -p migrations
	@TIMESTAMP=$$(date +%Y%m%d%H%M%S); \
	FILENAME="migrations/$${TIMESTAMP}_$(NAME).sql"; \
	echo "-- +goose Up" > $$FILENAME; \
	echo "" >> $$FILENAME; \
	echo "-- +goose Down" >> $$FILENAME; \
	echo "Created migration: $$FILENAME"

migration-rollback:
	@if [ -z "$(DB_DSN)" ]; then \
		echo "Error: DB_DSN is required"; \
		echo "Usage: make migration-rollback DB_DSN=<dsn>"; \
		exit 1; \
	fi
	@echo "Rolling back last migration..."
	goose -dir migrations postgres "$(DB_DSN)" down

generate-mocks:
	@mkdir -p internal/mocks
	mockgen -package=mocks -destination=internal/mocks/session_repository_mock.go -source=internal/database/session_repository_interface.go