<<<<<<< HEAD
.PHONY: migration migration-create migration-rollback
=======
.PHONY: migration migration-create generate-mocks
>>>>>>> 521f43f86c36cc74b0abe34cd15ecdca486ab2ca

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

<<<<<<< HEAD
migration-rollback:
	@if [ -z "$(DB_DSN)" ]; then \
		echo "Error: DB_DSN is required"; \
		echo "Usage: make migration-rollback DB_DSN=<dsn>"; \
		exit 1; \
	fi
	@echo "Rolling back last migration..."
	goose -dir migrations postgres "$(DB_DSN)" down
=======
generate-mocks:
	mkdir -p internal/mocks
		mockgen -package=mocks -destination=internal/mocks/session_repository_mock.go -source=internal/database/session_repository_interface.go

>>>>>>> 521f43f86c36cc74b0abe34cd15ecdca486ab2ca
