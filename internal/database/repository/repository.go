package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

type Repository struct {
	client *sqlx.DB
}

func NewRepository(client *sqlx.DB) *Repository {
	return &Repository{client: client}
}

func (r Repository) GetServiceByID(ctx context.Context, serviceID int) (models.Service, error) {
	return models.Service{}, nil
}

func (r Repository) CreateUser(ctx context.Context, telegramID int64, userName string, firstName string, lastName string) error {
	return nil
}
