package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

// RunMigrations применяет миграции и логирует результат
func RunMigrations(db *sql.DB) error {

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	// Логируем текущую версию БД
	versionBefore, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}
	log.Printf("Current database version: %d", versionBefore)

	// Применяем миграции
	err = goose.Up(db, "migrations")
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Логируем финальную версию и считаем применённые миграции
	versionAfter, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}

	// Количество применённых миграций = разница версий
	appliedCount := versionAfter - versionBefore
	log.Printf("Applied %d migration(s)", appliedCount)
	log.Printf("Final database version: %d", versionAfter)

	return nil
}
