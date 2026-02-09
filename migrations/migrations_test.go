package migrations

// Интеграционный тест миграции 002_create_bookings_table.sql.
// Проверяет критерии приёмки: миграция применяется, FK constraint работает, индексы созданы.

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBookingsMigration(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer db.Close()

	createUsersDependency(t, ctx, db)
	applyMigration(t, ctx, db, "002_create_bookings_table.sql")

	t.Run("Indexes", func(t *testing.T) {
		for _, idx := range []string{"idx_bookings_user_id", "idx_bookings_service_id", "idx_bookings_status"} {
			assert.True(t, indexExists(t, ctx, db, idx), "Индекс %s не создан", idx)
		}
	})

	t.Run("CascadeDelete", func(t *testing.T) {
		var userID int64
		err := db.QueryRowContext(ctx, `INSERT INTO users (telegram_id) VALUES (123456789) RETURNING id;`).Scan(&userID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `INSERT INTO bookings (user_id, service_id, booking_date, guest_name) VALUES ($1, 1, '2026-03-01', 'Test Guest');`, userID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, `DELETE FROM users WHERE id = $1;`, userID)
		require.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM bookings WHERE user_id = $1;`, userID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "CASCADE DELETE не сработал")
	})
}

func setupTestDatabase(t *testing.T, ctx context.Context) *sql.DB {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("Ошибка при остановке контейнера: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.PingContext(ctx))
	return db
}

func createUsersDependency(t *testing.T, ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
}

func applyMigration(t *testing.T, ctx context.Context, db *sql.DB, filename string) {
	migrationSQL, err := os.ReadFile(filepath.Join(".", filename))
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, string(migrationSQL))
	require.NoError(t, err)
}

func indexExists(t *testing.T, ctx context.Context, db *sql.DB, indexName string) bool {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT FROM pg_indexes WHERE indexname = $1);`, indexName).Scan(&exists)
	require.NoError(t, err)
	return exists
}
