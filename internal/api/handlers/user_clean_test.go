package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	repository "github.com/yandex-development-1-team/go/internal/repository/postgres"
	service "github.com/yandex-development-1-team/go/internal/service/api"
	"golang.org/x/crypto/bcrypt"
)

type seedUserParams struct {
	TelegramID int64
	Email      string
	Role       string
	Status     string
	PassHash   string
}

func setupServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()

	userRepo := repository.NewUserRepo(db)
	refreshRepo := repository.NewRefreshTokenRepo(db)
	svc := service.GetNewAuthService(userRepo, refreshRepo)
	handler := NewAuthHandler(svc)

	router := gin.New()
	router.POST("/auth/login", handler.HandleLogin)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func seedUser(t *testing.T, db *sqlx.DB, p seedUserParams) {
	t.Helper()
	_, err := db.Exec(`
			INSERT INTO users (telegram_id, email, role, status, password_hash)
			VALUES ($1, $2, $3, $4, $5)`,
		p.TelegramID, p.Email, p.Role, p.Status, p.PassHash,
	)
	require.NoError(t, err)
}

func TestHandleLogin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := startContainer()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, container.Terminate(ctx))
	}()

	db, err := createDB(container)
	require.NoError(t, err)
	defer db.Close()

	server := setupServer(t, db)

	validHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)

	// role — только 'admin' или 'manager', нет 'user'
	seedUser(t, db, seedUserParams{
		TelegramID: 1,
		Email:      "user@example.com",
		Status:     "active",
		Role:       "manager",
		PassHash:   string(validHash),
	})

	seedUser(t, db, seedUserParams{
		TelegramID: 2,
		Email:      "blocked@example.com",
		Status:     "blocked",
		Role:       "manager",
		PassHash:   string(validHash),
	})

	tests := []struct {
		name       string
		body       any
		wantStatus int
		checkBody  func(*testing.T, map[string]any)
	}{
		{
			name:       "успешный логин",
			body:       map[string]string{"login": "user@example.com", "password": "password123"},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["token"])
				assert.NotEmpty(t, body["refresh_token"])

				user, ok := body["user"].(map[string]any)
				require.True(t, ok, "user field missing or wrong type")
				assert.Equal(t, "user@example.com", user["email"])
				assert.Equal(t, "manager", user["role"])
				assert.Equal(t, "active", user["status"])
			},
		},
		{
			name:       "неверный пароль",
			body:       map[string]string{"login": "user@example.com", "password": "wrongpass"},
			wantStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "error", body["status"])
				assert.NotEmpty(t, body["message"])
			},
		},
		{
			name:       "пользователь не найден",
			body:       map[string]string{"login": "notfound@example.com", "password": "password123"},
			wantStatus: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "error", body["status"])
				assert.NotEmpty(t, body["message"])
			},
		},
		{
			name:       "заблокированный пользователь",
			body:       map[string]string{"login": "blocked@example.com", "password": "password123"},
			wantStatus: http.StatusForbidden,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "error", body["status"])
				assert.NotEmpty(t, body["message"])
			},
		},
		{
			name:       "короткий пароль",
			body:       map[string]string{"login": "user@example.com", "password": "123"},
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "error", body["status"])
			},
		},
		{
			name:       "пустой email",
			body:       map[string]string{"login": "", "password": "password123"},
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.Equal(t, "error", body["status"])
			},
		},
		{
			name:       "невалидный json",
			body:       "not json",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)

			resp, err := http.Post(
				server.URL+"/auth/login",
				"application/json",
				bytes.NewReader(body),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.checkBody != nil {
				var result map[string]any
				require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
				tt.checkBody(t, result)
			}
		})
	}
}

func startContainer() (tc.Container, error) {
	// настройка testcontainers postgres
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
					"host=%s port=%s user=postgres password=password dbname=postgres sslmode=disable connect_timeout=5",
					host, port.Port())
			}).
			WithStartupTimeout(120 * time.Second),
	}

	// генерация контейнера
	dbContainer, err := tc.GenericContainer(
		context.Background(),
		tc.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	if err != nil {
		fmt.Printf("Container logs: %v\n", err)
		return nil, err
	}

	return dbContainer, err
}

func createDB(container tc.Container) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf(
		"host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable",
		host, port.Port())

	db, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.UpContext(ctx, db.DB, "../../../migrations"); err != nil {
		return nil, err
	}

	return db, nil
}
