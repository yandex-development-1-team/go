package bot

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

const slugInfo = "org-info"

// AboutRepo defines the data access layer interface for service operations
type AboutRepo interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
}

// AboutService provides logic for service 'About us'
type AboutService struct {
	repo AboutRepo
}

// NewAboutService creates a new instance of the 'AboutService'
func NewAboutService(repo AboutRepo) *AboutService {
	return &AboutService{repo: repo}
}

// GetBySlug retrieves a resource page from the database by its slug
func (s *AboutService) GetBySlug(ctx context.Context) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slugInfo)
}
