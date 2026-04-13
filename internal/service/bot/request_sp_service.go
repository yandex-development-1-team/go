package bot

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

const slugRequestSp = "req-spec-projects"

// RequestSpRepo defines the data access layer interface for service operations
type RequestSpRepo interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
}

// RequestSpService provides logic for service 'guide'
type RequestSpService struct {
	repo GuideRepo
}

// NewRequestSpService creates a new instance of the 'RequestSpService'
func NewRequestSpService(repo GuideRepo) *RequestSpService {
	return &RequestSpService{repo: repo}
}

// GetBySlug retrieves a resource page from the database by its slug
func (s *RequestSpService) GetBySlug(ctx context.Context) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slugRequestSp)
}
