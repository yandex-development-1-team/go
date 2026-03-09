package service

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/repository/models"
	repository "github.com/yandex-development-1-team/go/internal/repository/postgres"
)

type SpecialProjectService struct {
	repo repository.SpecialProjectRepo
}

func NewSpecialProjectService(repo repository.SpecialProjectRepo) *SpecialProjectService {
	return &SpecialProjectService{repo: repo}
}

func (s *SpecialProjectService) UpdateSpecialProject(ctx context.Context, id int, specialProject models.SpecialProject) (models.SpecialProject, error) {
	if id <= 0 {
		return models.SpecialProject{}, models.ErrInvalidInput
	}
	if len(specialProject.Title) > 255 {
		return models.SpecialProject{}, models.ErrInvalidInput
	}
	return s.repo.UpdateSpecialProject(ctx, id, specialProject)
}

func (s *SpecialProjectService) DeleteSpecialProject(ctx context.Context, id int) error {
	if id <= 0 {
		return models.ErrInvalidInput
	}
	return s.repo.DeleteSpecialProject(ctx, id)
}
