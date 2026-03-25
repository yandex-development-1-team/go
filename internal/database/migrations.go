package database

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
)

func RunMigrations(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	versionBefore, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}
	logger.Info("Current database version",
		zap.Int64("version", versionBefore),
		zap.String("migrations_dir", migrationsDir),
	)
	migrations, err := goose.CollectMigrations(migrationsDir, 0, goose.MaxVersion)
	if err != nil {
		return fmt.Errorf("failed to collect migrations: %w", err)
	}

	pendingCount := 0
	for _, m := range migrations {
		if m.Version > versionBefore {
			pendingCount++
		}
	}

	if pendingCount > 0 {
		logger.Info("Found pending migrations",
			zap.Int("count", pendingCount),
		)

		if err := goose.Up(db, migrationsDir); err != nil {
			logger.Error("Failed to apply migrations",
				zap.Error(err))
			return fmt.Errorf("failed to apply migrations: %w", err)
		}

		versionAfter, err := goose.GetDBVersion(db)
		if err != nil {
			logger.Error("Failed to get DB version after migration",
				zap.Error(err),
			)
			return fmt.Errorf("failed to get DB version: %w", err)
		}

		logger.Info("Migrations applied successfully",
			zap.Int("applied_count", pendingCount),
			zap.Int64("version_before", versionBefore),
			zap.Int64("version_after", versionAfter),
		)
	} else {
		logger.Info("Database is up to date",
			zap.Int64("version", versionBefore),
		)
	}

	return nil
}
