package bot

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

const slugLinks = "useful-links"

// UsefulLinksRepo defines the data access layer interface for service operations
type UsefulLinksRepo interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
}

// UsefulLinksService provides logic for service 'guide'
type UsefulLinksService struct {
	repo UsefulLinksRepo
}

// NewUsefulLinksService creates a new instance of the 'UsefulLinksService'
func NewUsefulLinksService(repo GuideRepo) *UsefulLinksService {
	return &UsefulLinksService{repo: repo}
}

// GetBySlug retrieves a resource page from the database by its slug
func (s *UsefulLinksService) GetBySlug(ctx context.Context) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slugLinks)
}
