package bot

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

const slugExamplesSp = "spec-projects"

// ExamplesSpRepo defines the data access layer interface for service operations
type ExamplesSpRepo interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
}

// ExamplesSpService provides logic for service 'guide'
type ExamplesSpService struct {
	repo ExamplesSpRepo
}

// NewExamplesSpService creates a new instance of the 'GuideService'
func NewExamplesSpService(repo GuideRepo) *ExamplesSpService {
	return &ExamplesSpService{repo: repo}
}

// GetBySlug retrieves a resource page from the database by its slug
func (s *ExamplesSpService) GetBySlug(ctx context.Context) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slugExamplesSp)
}
