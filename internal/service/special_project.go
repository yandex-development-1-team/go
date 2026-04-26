package service

import (
	"context"
	"errors"

	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository"
)

type SpecialProjectService struct {
	repo repository.SpecialProjectRepository
}

func NewSpecialProjectService(repo repository.SpecialProjectRepository) *SpecialProjectService {
	return &SpecialProjectService{repo: repo}
}

func (s *SpecialProjectService) Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error) {
	if proj.Title == "" {
		return nil, errors.New("title is required")
	}

	dbModel, err := s.repo.Create(ctx, proj)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectAlreadyExists) {
			return nil, models.ErrSpecialProjectAlreadyExists
		}
		return nil, err
	}
	return dbModel, nil
}

func (s *SpecialProjectService) GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
	dbModel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectNotFound) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, err
	}
	return dbModel, nil
}

func (s *SpecialProjectService) List(ctx context.Context, statusStr string, search string, limit, offset int) ([]*models.SpecialProjectDB, int, error) {

	dbList, total, err := s.repo.List(ctx, statusStr, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return dbList, total, nil
}

func (s *SpecialProjectService) Update(ctx context.Context, id int64, proj *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
	if id <= 0 || proj == nil {
		return nil, models.ErrInvalidInput
	}
	dbModel, err := s.repo.Update(ctx, id, proj)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectNotFound) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, err
	}
	return dbModel, nil
}

func (s *SpecialProjectService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return models.ErrInvalidInput
	}
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectNotFound) {
			return models.ErrSpecialProjectNotFound
		}
		return err
	}
	return nil
}
