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

.PHONY: migration migration-create generate-mocks

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
	
## test: running tests
test:
	$(GOTEST) ./... -v -cover -count=1

## fmt: Format the source code
fmt:
	$(GO) fmt ./...

## vet: Check the code for suspicious structures
vet:
	$(GO) vet ./...

## lint: Launch the linter (golangci-lint)
lint:
	golangci-lint run ./...


migration:
	@echo "Migration commands:"
	@echo "  make migration-create NAME=<name>  Create new migration file"
	@echo ""
	@echo "Example:"
	@echo "  make migration-create NAME=create_users_table"

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

generate-mocks:
	mkdir -p internal/mocks
		mockgen -package=mocks -destination=internal/mocks/session_repository_mock.go -source=internal/database/session_repository_interface.go

