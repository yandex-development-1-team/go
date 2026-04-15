////go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database"
	pgrepo "github.com/yandex-development-1-team/go/internal/repository/postgres"
	service "github.com/yandex-development-1-team/go/internal/service/api"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := startContainer()
	if err != nil {
		log.Fatal(err)
	}

	db, err = createDB(container)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err.Error())
	}

	code := m.Run()

	_ = container.Terminate(ctx)
	os.Exit(code)
}

type seedUserParams struct {
	TelegramNick string
	FirstName    string
	LastName     string
	Email        string
	Role         string
	Status       string
	PassHash     string
}

func setupServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()

	userRepo := pgrepo.NewStaffRepo(db)
	refreshRepo := pgrepo.NewRefreshTokenRepo(db)
	passwordResetRepo := pgrepo.NewPasswordResetRepository(db)
	txRepo := pgrepo.NewTxRepo(db)
	emailService := service.NewEmailService(config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     1025,
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "test@example.com",
		BaseURL:      "http://localhost",
	})
	svc := service.NewAuthService(db, refreshRepo, passwordResetRepo, userRepo, emailService, txRepo, "test-secret", 15, 30)
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
		INSERT INTO staff (telegram_nick, first_name, last_name, email, role, status, password_hash)
		VALUES ($1, $2, $3, $4, $5::user_role_type, $6::user_status_type, $7)`,
		p.TelegramNick, p.FirstName, p.LastName, p.Email, p.Role, p.Status, p.PassHash,
	)
	require.NoError(t, err)
}

func TestHandleLogin(t *testing.T) {
	_, err := db.Exec(`TRUNCATE TABLE staff CASCADE`)

	server := setupServer(t, db)

	validHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)

	seedUser(t, db, seedUserParams{
		TelegramNick: "nick1",
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "user@example.com",
		Status:       "active",
		Role:         "manager_1",
		PassHash:     string(validHash),
	})

	seedUser(t, db, seedUserParams{
		TelegramNick: "nick2",
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "blocked@example.com",
		Status:       "blocked",
		Role:         "manager_1",
		PassHash:     string(validHash),
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
				assert.Equal(t, "manager_1", user["role"])
				assert.Equal(t, "active", user["status"])
			},
		},
		{
			name:       "неверный пароль",
			body:       map[string]string{"login": "user@example.com", "password": "wrongpass"},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name:       "пользователь не найден",
			body:       map[string]string{"login": "notfound@example.com", "password": "password123"},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name:       "заблокированный пользователь",
			body:       map[string]string{"login": "blocked@example.com", "password": "password123"},
			wantStatus: http.StatusForbidden,
			checkBody:  checkServiceErrorBody,
		},
		{
			name:       "короткий пароль",
			body:       map[string]string{"login": "user@example.com", "password": "123"},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name:       "пустой email",
			body:       map[string]string{"login": "", "password": "password123"},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
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

func TestRegisterHandler(t *testing.T) {
	_, err := db.Exec(`TRUNCATE TABLE staff CASCADE`)

	server := setupRegisterServer(t, db)

	validHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)

	seedStaff(t, db, "existing@example.com", string(validHash))

	const inviteOK = "invite-token-ok"

	tests := []struct {
		name       string
		body       any
		wantStatus int
		checkBody  func(*testing.T, map[string]any)
	}{
		{
			name: "успешная регистрация",
			body: map[string]string{
				"first_name":   "New User",
				"last_name":    "User",
				"email":        "newuser@example.com",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, body map[string]any) {
				assert.NotEmpty(t, body["token"])
				assert.NotEmpty(t, body["refresh_token"])

				user, ok := body["user"].(map[string]any)
				require.True(t, ok, "user field missing or wrong type")
				assert.Equal(t, "newuser@example.com", user["email"])
				assert.Equal(t, "invited", user["status"])
			},
		},
		{
			name: "email уже существует",
			body: map[string]string{
				"first_name":   "Test User",
				"last_name":    "User",
				"email":        "existing@example.com",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusConflict,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "пустое имя",
			body: map[string]string{
				"first_name":   "",
				"last_name":    "User",
				"email":        "test@example.com",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "имя короче 2 символов",
			body: map[string]string{
				"first_name":   "A",
				"last_name":    "User",
				"email":        "test@example.com",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "пустой email",
			body: map[string]string{
				"first_name":   "Test User",
				"last_name":    "User",
				"email":        "",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "невалидный email",
			body: map[string]string{
				"first_name":   "Test User",
				"last_name":    "User",
				"email":        "not-an-email",
				"password":     "password123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "пароль короче 8 символов",
			body: map[string]string{
				"first_name":   "Test User",
				"last_name":    "User",
				"email":        "test2@example.com",
				"password":     "123",
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
		},
		{
			name: "пароль длиннее 72 символов",
			body: map[string]string{
				"first_name":   "Test User",
				"last_name":    "User",
				"email":        "test3@example.com",
				"password":     strings.Repeat("a", 73),
				"invite_token": inviteOK,
			},
			wantStatus: http.StatusBadRequest,
			checkBody:  checkServiceErrorBody,
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
				server.URL+"/auth/register",
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

func checkServiceErrorBody(t *testing.T, body map[string]any) {
	errors, ok := body["errors"].([]any)
	require.True(t, ok, "expected 'errors' array in body: %v", body)
	assert.NotEmpty(t, errors, "errors must not be empty")
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
					"host=%s port=%s user=postgres password=password dbname=postgres sslmode=disable connect_timeout=5",
					host, port.Port())
			}).
			WithStartupTimeout(120 * time.Second),
	}

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
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf(
		"host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable",
		host, port.Port())

	db, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	migDir, err := database.ResolveMigrationsDir("")
	if err != nil {
		return nil, err
	}
	if err := database.RunMigrations(db.DB, migDir); err != nil {
		return nil, err
	}

	return db, nil
}

func setupRegisterServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()

	userRepo := pgrepo.NewStaffRepo(db)
	refreshRepo := pgrepo.NewRefreshTokenRepo(db)
	passwordResetRepo := pgrepo.NewPasswordResetRepository(db)
	txRepo := pgrepo.NewTxRepo(db)
	emailService := service.NewEmailService(config.EmailConfig{
		SMTPHost:     "localhost",
		SMTPPort:     1025,
		SMTPUsername: "",
		SMTPPassword: "",
		FromEmail:    "test@example.com",
		BaseURL:      "http://localhost",
	})
	svc := service.NewAuthService(db, refreshRepo, passwordResetRepo, userRepo, emailService, txRepo, "test-secret", 15, 30)
	handler := NewAuthHandler(svc)

	router := gin.New()
	router.POST("/auth/register", handler.RegisterHandler)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func seedStaff(t *testing.T, db *sqlx.DB, email string, passHash string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO staff (first_name, last_name, email, password_hash, role, status)
		VALUES ($1, $2, $3, $4, $5::user_role_type, $6::user_status_type)`,
		"Existing", "User", email, passHash, "manager_1", "active",
	)
	require.NoError(t, err)
}
