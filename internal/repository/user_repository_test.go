package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var (
	db     *sqlx.DB
	logger *zap.Logger
	repo   UserRepository
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	container, err := startContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = createDB(container)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err.Error())
	}

	logger = nopLogger()
	repo = NewUserRepository(db, logger)

	code := m.Run()

	err = container.Terminate(ctx)
	if err != nil {
		log.Fatalf("failed to terminate db container: %s", err.Error())
	}

	os.Exit(code)
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

func createDB(container tc.Container) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf(
		"host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable",
		host, port.Port())
	db, err = sqlx.Connect("postgres", dbURI)
	if err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	_, err = goose.GetDBVersion(db.DB)
	if !errors.Is(err, goose.ErrNoMigrations) {
		_, err := db.Exec("DROP TABLE IF EXISTS goose_db_version")
		if err != nil {
			return err
		}
	}

	if err := goose.UpContext(ctx, db.DB, "../../migrations"); err != nil {
		return err
	}

	_, err = goose.GetDBVersion(db.DB)
	if errors.Is(err, goose.ErrNoMigrations) {
		return err
	}

	return nil
}

// NopLogger - логгер который ничего не делает
func nopLogger() *zap.Logger {
	return zap.NewNop()
}

func TestCreateUser(t *testing.T) {

	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		reqUserName     string
		reqFirstName    string
		reqLastName     string
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "double_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqUserName:     "",
			reqFirstName:    "",
			reqLastName:     "",
			wantErr:         ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   123456,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         ErrRequestTimeout,
		},
	}

	err := db.Ping()
	if err != nil {
		log.Println("Нет соединения с базой данных:", err)
	} else {
		log.Println("Соединение с базой успешно установлено")
	}

	for _, tt := range tests {

		ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
		defer cancel()

		// тест отмены контекста
		if tt.name == "request_canceled" {
			cancel()
		}

		err := repo.CreateUser(ctx, tt.reqTelegramId, tt.reqUserName, tt.reqFirstName, tt.reqLastName)
		assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		wantUserId      int64
		wantUserName    string
		wantFirstName   string
		wantLastName    string
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantUserId:      1,
			wantUserName:    "test_name",
			wantFirstName:   "test_first_name",
			wantLastName:    "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "user_not_found",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantErr:         ErrUserNotFound,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantErr:         ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   123456,
			wantErr:         ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
		defer cancel()

		if tt.name == "request_canceled" {
			cancel()
		}

		user, err := repo.GetUserByTelegramID(ctx, tt.reqTelegramId)
		if tt.wantErr != nil {
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		}
		if tt.wantErr == nil {
			assert.NoError(t, err, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
			assert.Equal(t, tt.wantUserId, user.ID, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
			assert.Equal(t, tt.wantFirstName, user.FirstName, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
			assert.Equal(t, tt.wantFirstName, user.FirstName, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
			assert.Equal(t, tt.wantLastName, user.LastName, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestUpdateUserGrade(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		reqGrade        int
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         nil,
		},
		{
			name:            "wrong_user",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			reqGrade:        2,
			wantErr:         ErrUserNotFound,
		},
		{
			name:            "context_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         ErrRequestCanceled,
		},
		{
			name:            "context_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
		defer cancel()

		if tt.name == "context_canceled" {
			cancel()
		}

		err := repo.UpdateUserGrade(ctx, tt.reqTelegramId, tt.reqGrade)
		assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		wantAdmin       bool
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantAdmin:       true,
			wantErr:         nil,
		},
		{
			name:            "wrong_user",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         ErrUserNotFound,
		},
		{
			name:            "context_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         ErrRequestCanceled,
		},
		{
			name:            "context_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         ErrRequestCanceled,
		},
	}

	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
		defer cancel()

		if tt.name == "context_canceled" {
			cancel()
		}

		isAdmin, err := repo.IsAdmin(ctx, tt.reqTelegramId)
		if tt.wantErr == nil {
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		} else {
			assert.Equal(t, tt.wantAdmin, isAdmin, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		}
	}
}
