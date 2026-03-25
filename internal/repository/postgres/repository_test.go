package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
)

var (
	db          *sqlx.DB
	repoSession *pgSessionRepo
)

func TestMain(m *testing.M) {
	logger.NewLogger("dev", "debug")
	metrics.Initialize(config.Config{Environment: "test", HostName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := startContainer()
	if err != nil {
		log.Fatal(err)
	}

	db, err = setupTestDB(container)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err.Error())
	}

	repoSession = NewSessionRepository(db)

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
		WaitingFor: wait.ForSQL(nat.Port("5432/tcp"), "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(120 * time.Second),
	}

	return tc.GenericContainer(context.Background(), tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func setupTestDB(container tc.Container) (*sqlx.DB, error) {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())

	newdb, err := sqlx.Connect("postgres", dbURI)
	if err != nil {
		return nil, err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.UpContext(context.Background(), newdb.DB, "../../../migrations"); err != nil {
		return nil, err
	}

	return newdb, nil
}
