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

func toDomain(db *models.SpecialProjectDB) *models.SpecialProject {
	if db == nil {
		return nil
	}
	return &models.SpecialProject{
		ID:          db.ID,
		Title:       db.Title,
		Description: db.Description,
		Image:       db.Image,
		Status:      db.Status,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func fromDomain(domain *models.SpecialProject) *models.SpecialProjectDB {
	if domain == nil {
		return nil
	}
	return &models.SpecialProjectDB{
		ID:          domain.ID,
		Title:       domain.Title,
		Description: domain.Description,
		Image:       domain.Image,
		Status:      domain.Status,
	}
}

func (s *SpecialProjectService) Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProject, error) {
	if proj.Title == "" {
		return nil, errors.New("title is required")
	}
	dbModel := fromDomain(proj)
	dbModel, err := s.repo.Create(ctx, dbModel)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectAlreadyExists) {
			return nil, models.ErrSpecialProjectAlreadyExists
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) GetByID(ctx context.Context, id int64) (*models.SpecialProject, error) {
	dbModel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectNotFound) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) List(ctx context.Context, statusStr string, search string, limit, offset int) ([]*models.SpecialProject, int, error) {

	dbList, total, err := s.repo.List(ctx, statusStr, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*models.SpecialProject, 0, len(dbList))
	for _, item := range dbList {
		result = append(result, &models.SpecialProject{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Image:       item.Image,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return result, total, nil
}

func (s *SpecialProjectService) Update(ctx context.Context, id int64, proj *models.SpecialProject) (*models.SpecialProject, error) {
	if id <= 0 || proj == nil {
		return nil, models.ErrInvalidInput
	}
	if len(proj.Title) > 255 {
		return nil, models.ErrInvalidInput
	}
	update := &models.SpecialProjectUpdate{
		Title:       proj.Title,
		Description: proj.Description,
		Image:       proj.Image,
		Status:      proj.Status,
	}
	dbModel, err := s.repo.Update(ctx, id, update)
	if err != nil {
		if errors.Is(err, models.ErrSpecialProjectNotFound) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, err
	}
	return toDomain(dbModel), nil
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
