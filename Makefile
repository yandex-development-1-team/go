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

.PHONY: migration migration-create migration-rollback generate-mocks

all: build

## build: building a project
build:
	$(GOBUILD) -C $(SRC_DIR) -o $(BUILD_DIR)/$(BIN_NAME)

## run: run the application
run:
	$(GORUN) $(SRC_DIR)/*.go

## clean: delete binary
clean:	
	$(GOCLEAN)
	rm -rf $(SRC_DIR)/bin
	
## test: running unit tests (без интеграционных)
test:
	$(GOTEST) ./... -v -count=1

## test-integration: интеграционные тесты (testcontainers: Postgres и т.д.). То же вызывается в CI.
test-integration:
	$(GOTEST) -tags=integration ./... -v -count=1 -timeout=15m

# Local module path for goimports grouping
IMPORT_LOCAL=github.com/yandex-development-1-team/go

GO_FILES=$(shell find . -name '*.go' -not -path './bot/*' 2>/dev/null)

## fmt: Format the source code (gofmt)
fmt:
	@gofmt -w $(GO_FILES)

## imports: Sort and group imports (goimports)
imports:
	@$(GORUN) golang.org/x/tools/cmd/goimports@latest -w -local $(IMPORT_LOCAL) $(GO_FILES)

## vet: Check the code for suspicious structures
vet:
	$(GO) vet ./...

## lint: Run golangci-lint (нужна сборка под Go 1.26; в CI используется action с нужной версией)
lint:
	golangci-lint run ./... --timeout=5m

## check: Verify formatting and imports (CI check, fails if not applied)
check: check-fmt check-imports vet
	@echo "check: fmt, imports, vet OK"

check-fmt:
	@out=$$(gofmt -l $(GO_FILES)); [ -z "$$out" ] || { echo "gofmt -l: unformatted files:"; echo "$$out"; exit 1; }

check-imports:
	@out=$$($(GORUN) golang.org/x/tools/cmd/goimports@latest -l -local $(IMPORT_LOCAL) $(GO_FILES)); [ -z "$$out" ] || { echo "goimports -l: files with import issues:"; echo "$$out"; exit 1; }


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