package service

import (
	"context"
	"errors"

	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/specialproject"
)

type SpecialProjectService struct {
	repo repository.SpecialProjectRepository
}

func NewSpecialProjectService(repo repository.SpecialProjectRepository) *SpecialProjectService {
	return &SpecialProjectService{repo: repo}
}

func toDomain(db *specialproject.DB) *specialproject.Project {
	if db == nil {
		return nil
	}
	return &specialproject.Project{
		ID:            db.ID,
		Title:         db.Title,
		Description:   db.Description,
		Image:         db.Image,
		IsActiveInBot: db.IsActiveInBot,
		CreatedAt:     db.CreatedAt,
		UpdatedAt:     db.UpdatedAt,
	}
}

func fromDomain(domain *specialproject.Project) *specialproject.DB {
	if domain == nil {
		return nil
	}
	return &specialproject.DB{
		ID:            domain.ID,
		Title:         domain.Title,
		Description:   domain.Description,
		Image:         domain.Image,
		IsActiveInBot: domain.IsActiveInBot,
	}
}

func (s *SpecialProjectService) Create(ctx context.Context, proj *specialproject.Project) (*specialproject.Project, error) {
	if proj.Title == "" {
		return nil, errors.New("title is required")
	}
	dbModel := fromDomain(proj)
	dbModel, err := s.repo.Create(ctx, dbModel)
	if err != nil {
		if errors.Is(err, specialproject.ErrAlreadyExists) {
			return nil, specialproject.ErrAlreadyExists
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) GetByID(ctx context.Context, id int64) (*specialproject.Project, error) {
	dbModel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, specialproject.ErrNotFound) {
			return nil, specialproject.ErrNotFound
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) List(ctx context.Context, statusStr string, search string) ([]specialproject.Project, error) {
	var statusFilter *bool
	if statusStr != "" {
		val := statusStr == "active"
		statusFilter = &val
	}
	dbList, err := s.repo.List(ctx, statusFilter, search)
	if err != nil {
		return nil, err
	}
	result := make([]specialproject.Project, 0, len(dbList))
	for _, item := range dbList {
		result = append(result, specialproject.Project{
			ID:            item.ID,
			Title:         item.Title,
			IsActiveInBot: item.IsActiveInBot,
		})
	}
	return result, nil
}

func (s *SpecialProjectService) UpdateSpecialProject(ctx context.Context, id int64, proj *specialproject.Project) (*specialproject.Project, error) {
	if id <= 0 || proj == nil {
		return nil, models.ErrInvalidInput
	}
	if len(proj.Title) > 255 {
		return nil, models.ErrInvalidInput
	}
	update := &specialproject.Update{
		Title:         proj.Title,
		Description:   proj.Description,
		Image:         proj.Image,
		IsActiveInBot: proj.IsActiveInBot,
	}
	dbModel, err := s.repo.UpdateSpecialProject(ctx, id, update)
	if err != nil {
		if errors.Is(err, specialproject.ErrNotFound) {
			return nil, specialproject.ErrNotFound
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) DeleteSpecialProject(ctx context.Context, id int64) error {
	if id <= 0 {
		return models.ErrInvalidInput
	}
	err := s.repo.DeleteSpecialProject(ctx, id)
	if err != nil {
		if errors.Is(err, specialproject.ErrNotFound) {
			return specialproject.ErrNotFound
		}
		return err
	}
	return nil
}
