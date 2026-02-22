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

	// Получаем текущую версию БД
	versionBefore, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get DB version: %w", err)
	}
	log.Printf("Current database version: %d", versionBefore)

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
		log.Printf("Found %d pending migration(s)", pendingCount)

		if err := goose.Up(db, "migrations"); err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}

		// Получаем новую версию
		versionAfter, err := goose.GetDBVersion(db)
		if err != nil {
			return fmt.Errorf("failed to get DB version: %w", err)
		}

		// Логируем точное количество применённых миграций
		log.Printf("Applied %d migration(s)", pendingCount)
		log.Printf("Database version: %d → %d", versionBefore, versionAfter)
	} else {
		log.Printf("Database is up to date (version: %d)", versionBefore)
	}

	return nil
}
