package service

import (
	"context"
	"fmt"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type ResourcePageService struct {
	repo repository.ResourcePageRepository
}

func NewResourcePageService(repo repository.ResourcePageRepository) *ResourcePageService {
	return &ResourcePageService{repo: repo}
}

func (s *ResourcePageService) GetResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *ResourcePageService) GetAllResourcePages(ctx context.Context) ([]models.ResourcePage, error) {
	return s.repo.GetAll(ctx)
}

func (s *ResourcePageService) UpdateResourcePage(ctx context.Context, slug string, page models.ResourcePage) (*models.ResourcePage, error) {
	updated, err := s.repo.Update(ctx, slug, page)
	if err != nil {
		return nil, fmt.Errorf("update resource page: %w", err)
	}

	return updated, nil
}

func (s *ResourcePageService) ClearResourcePage(ctx context.Context, slug string) (*models.ResourcePage, error) {
	cleared, err := s.repo.Clear(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("clear resource page: %w", err)
	}

	return cleared, nil
}

func (s *ResourcePageService) DeleteResourcePageLink(ctx context.Context, slug string, id string) (*models.ResourcePage, error) {
	page, err := s.repo.DeleteLink(ctx, slug, id)
	if err != nil {
		return nil, fmt.Errorf("delete resource page link: %w", err)
	}
	return page, nil
}
