package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	dbmodels "github.com/yandex-development-1-team/go/internal/database/repository/models"
)

type Repository struct {
	client *sqlx.DB
}

func NewRepository(client *sqlx.DB) *Repository {
	return &Repository{client: client}
}

func (r Repository) GetServiceByID(ctx context.Context, serviceID int) (dbmodels.Service, error) {
	return dbmodels.Service{}, nil
}

func (r Repository) CreateUser(ctx context.Context, telegramID int64, userName string, firstName string, lastName string) error {
	return nil
}
