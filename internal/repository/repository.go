package repository

import (
	"context"
	dbmodels "github.com/yandex-development-1-team/go/internal/repository/models"
)

type Repository struct {
	//todo gorm client
}

func NewRepository() Repository {
	return Repository{}
}

func (r Repository) GetServiceByID(ctx context.Context, serviceID int) (dbmodels.Service, error) {
	return dbmodels.Service{}, nil
}

func (r Repository) CreateUser(ctx context.Context, telegramID int64, userName string, firstName string, lastName string) error {
	return nil
}

func (r Repository) GetBoxSolutions(ctx context.Context) ([]dbmodels.BoxSolution, error) {
	return nil, nil
}
