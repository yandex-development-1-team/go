package repository

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

type SpecialProjectRepository interface {
	Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error)
	GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	List(ctx context.Context, statusFilter *bool, searchQuery string) ([]*models.SpecialProjectDB, error)
}
