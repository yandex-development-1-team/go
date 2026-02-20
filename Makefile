.PHONY: migration migration-create generate-mocks

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

