.PHONY: migration migration-help

migration-help:
	@echo "Migration commands:"
	@echo "  make migration name=<migration_name>  Create new migration with timestamp"
	@echo "  make migration name=create_users_table"
	@echo ""
	@echo "Example:"
	@echo "  make migration name=add_email_column"
migration:
	@if [ -z "$(name)" ]; then \
		echo "Error: migration name is required"; \
		echo "Usage: make migration name=<migration_name>"; \
		echo ""; \
		make migration-help; \
		exit 1; \
	fi
	@mkdir -p migrations
	@TIMESTAMP=$$(date +%Y%m%d%H%M%S); \
	FILENAME="migrations/$${TIMESTAMP}_$(name).sql"; \
	echo "-- +goose Up" > $$FILENAME; \
	echo "" >> $$FILENAME; \
	echo "-- +goose Down" >> $$FILENAME; \
	echo "Created migration: $$FILENAME"
	@echo "   Edit the file to add SQL commands"
