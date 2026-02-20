.PHONY: generate-mocks lint lint-fix

generate-mocks:
	mkdir -p internal/mocks
		mockgen -package=mocks -destination=internal/mocks/session_repository_mock.go -source=internal/database/session_repository_interface.go
