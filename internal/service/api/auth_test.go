package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	pgrepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
)

var (
	db       *sqlx.DB
	rtRepo   *pgrepo.RefreshTokenRepo
	userRepo *pgrepo.UserRepo
	svc      *AuthService
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := startContainer()
	if err != nil {
		log.Fatal(err)
	}

	if err := createDB(container); err != nil {
		log.Fatalf("failed to init db: %s", err.Error())
	}
	rtRepo = pgrepo.NewRefreshTokenRepo(db)
	userRepo = pgrepo.NewUserRepo(db)
	txRepo := pgrepo.NewTxRepo(db)
	svc = NewAuthService(
		db,
		rtRepo,
		userRepo,
		txRepo,
		"test-service",
		15,
		7,
	)

	code := m.Run()

	_ = container.Terminate(ctx)
	os.Exit(code)
}

func startContainer() (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL(
			nat.Port("5432/tcp"),
			"postgres",
			func(host string, port nat.Port) string {
				return fmt.Sprintf(
					"host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable connect_timeout=5",
					host, port.Port())
			}).WithStartupTimeout(120 * time.Second),
	}

	return tc.GenericContainer(
		context.Background(),
		tc.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
}

func createDB(container tc.Container) error {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf(
		"host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable",
		host, port.Port())

	var err error
	db, err = sqlx.Connect("postgres", dbURI)
	if err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.UpContext(context.Background(), db.DB, "../../../migrations"); err != nil {
		return err
	}

	return nil
}

func insertTestUser(t *testing.T) int64 {
	t.Helper()

	tgID := time.Now().UnixNano()
	email := fmt.Sprintf("auth_tester_%d@test.local", tgID)

	var id int64
	err := db.QueryRow(`
	INSERT INTO staff (email, password_hash, role, status)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (email) DO UPDATE SET email=EXCLUDED.email
	RETURNING id`,
		email, "placeholder_hash", "manager", "active",
	).Scan(&id)
	assert.NoError(t, err)
	return id
}

func clearRefreshTokens(t *testing.T) {
	t.Helper()
	_, err := db.Exec("DELETE FROM refresh_tokens")
	assert.NoError(t, err)
}

func TestAuthService_Refresh_Success(t *testing.T) {
	clearRefreshTokens(t)
	userID := insertTestUser(t)

	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`,
		userID,
		"valid-refresh-token",
		time.Now().Add(24*time.Hour),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	accessToken, err := svc.Refresh(ctx, "valid-refresh-token")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken, "access token should not be empty")

	var revokedAt *time.Time
	err = db.Get(&revokedAt, `SELECT revoked_at FROM refresh_tokens WHERE token = $1`, "valid-refresh-token")
	assert.NoError(t, err)
	assert.NotNil(t, revokedAt, "old refresh token should be revoked")

	var count int
	err = db.Get(&count, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1`, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "should be two refresh tokens: old revoked + new active")
}

func TestAuthService_Refresh_Expired(t *testing.T) {
	clearRefreshTokens(t)
	userID := insertTestUser(t)

	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`,
		userID,
		"expired-refresh-token",
		time.Now().Add(-1*time.Hour),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = svc.Refresh(ctx, "expired-refresh-token")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pgrepo.ErrRefreshTokenExpired))
}

func TestAuthService_Refresh_Revoked(t *testing.T) {
	clearRefreshTokens(t)
	userID := insertTestUser(t)

	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at, revoked_at)
		VALUES ($1, $2, $3, NOW())
		`,
		userID,
		"revoked-refresh-token",
		time.Now().Add(24*time.Hour),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = svc.Refresh(ctx, "revoked-refresh-token")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pgrepo.ErrRefreshTokenRevoked))
}

func TestAuthService_Refresh_NotFound(t *testing.T) {
	clearRefreshTokens(t)
	_ = insertTestUser(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := svc.Refresh(ctx, "unknown-token")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, pgrepo.ErrRefreshTokenNotFound))
}

func TestAuthService_Refresh_ConcurrentRotation(t *testing.T) {
	clearRefreshTokens(t)
	userID := insertTestUser(t)

	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`,
		userID,
		"race-refresh-token",
		time.Now().Add(24*time.Hour),
	)
	assert.NoError(t, err)

	const goroutines = 10
	results := make(chan error, goroutines)
	var successCount int32

	for i := 0; i < goroutines; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := svc.Refresh(ctx, "race-refresh-token")
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			}
			results <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-results
	}

	assert.Equal(t, int32(1), successCount, "only one refresh should succeed")

	var revokedAt *time.Time
	err = db.Get(&revokedAt, `SELECT revoked_at FROM refresh_tokens WHERE token = $1`, "race-refresh-token")
	assert.NoError(t, err)
	assert.NotNil(t, revokedAt)

	var activeCount int
	err = db.Get(&activeCount, `
		SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, activeCount, 1)
}

func TestAuthService_Logout(t *testing.T) {
	clearRefreshTokens(t)
	userID := insertTestUser(t)

	_, err := db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		`,
		userID,
		"logout-refresh-token",
		time.Now().Add(24*time.Hour),
	)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = svc.Logout(ctx, "logout-refresh-token")
	assert.NoError(t, err)

	var revokedAt *time.Time
	err = db.Get(&revokedAt, `SELECT revoked_at FROM refresh_tokens WHERE token = $1`, "logout-refresh-token")
	assert.NoError(t, err)
	assert.NotNil(t, revokedAt)

	err = svc.Logout(ctx, "logout-refresh-token")
	assert.Error(t, err)
}
