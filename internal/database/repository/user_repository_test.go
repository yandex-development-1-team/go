package repository

// Интеграционные тесты репозитория пользователей (testcontainers + Postgres)
// находятся в user_repository_integration_test.go и запускаются только с тегом integration:
//
//	go test -tags=integration ./internal/database/repository/...
//	make test-integration
