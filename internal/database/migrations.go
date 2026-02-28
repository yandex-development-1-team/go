package database

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/yandex-development-1-team/go/internal/logger"
)

// RunMigrations применяет миграции и логирует результат
func RunMigrations(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Получаем текущую версию БД
	versionBefore, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}
	logger.Info("Current database version",
		zap.Int("version", versionBefore),
	)
	// Собираем все миграции для подсчёта ожидающих
	migrations, err := goose.CollectMigrations("migrations", 0, goose.MaxVersion)
	if err != nil {
		return fmt.Errorf("failed to collect migrations: %w", err)
	}

	// Считаем количество миграций, которые новее текущей версии
	pendingCount := 0
	for _, m := range migrations {
		if m.Version > versionBefore {
			pendingCount++
		}
	}

	// Если есть ожидающие миграции — применяем и логируем
	if pendingCount > 0 {
		logger.Info("Found pending migrations",
			zap.Int("count", pendingCount),
		)

		if err := goose.Up(db, "migrations"); err != nil {
			logger.Error("Failed to apply migrations",
				zap.Error(err),
			)
			return fmt.Errorf("failed to apply migrations: %w", err)
		}

		// Получаем новую версию
		versionAfter, err := goose.GetDBVersion(db)
		if err != nil {
			logger.Error("Failed to get DB version after migration",
				zap.Error(err),
			)
			return fmt.Errorf("failed to get DB version: %w", err)
		}

		// Логируем точное количество применённых миграций
		logger.Info("Migrations applied successfully",
			zap.Int("applied_count", pendingCount),
			zap.Int("version_before", versionBefore),
			zap.Int("version_after", versionAfter),
		)
	} else {
		logger.Info("Database is up to date",
			zap.Int("version", versionBefore),
		)
	}

	return nil
}
