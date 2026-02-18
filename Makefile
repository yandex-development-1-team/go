.PHONY: migration-create

migration-create: ## Create migration with timestamp
	@test -n "$(NAME)" || (echo "Usage: make migration-create NAME=migration_name" && exit 1)
	@TIMESTAMP=$$(date +%Y%m%d%H%M%S); \
	FILENAME="migrations/$${TIMESTAMP}_$(NAME).sql"; \
	echo "-- Migration: $${TIMESTAMP}_$(NAME)\n\n-- SQL code here\n" > $$FILENAME; \
	echo "Created: $$FILENAME"
