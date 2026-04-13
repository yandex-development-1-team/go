package bot

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

const slugGuide = "faq"

// GuideRepo defines the data access layer interface for service operations
type GuideRepo interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
}

// GuideService provides logic for service 'guide'
type GuideService struct {
	repo GuideRepo
}

// NewGuideService creates a new instance of the 'GuideService'
func NewGuideService(repo GuideRepo) *GuideService {
	return &GuideService{repo: repo}
}

// GetBySlug retrieves a resource page from the database by its slug
func (s *GuideService) GetBySlug(ctx context.Context) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slugGuide)
}
