package service

import (
	"context"
	"errors"
	"time"

	repository "github.com/yandex-development-1-team/go/internal/repository"
	dbmodels "github.com/yandex-development-1-team/go/internal/repository/models"
	"github.com/yandex-development-1-team/go/internal/service/models"
)

var (
	ErrNotFound = errors.New("special project not found")
)

type SpecialProjectService struct {
	repo repository.SpecialProjectRepository
}

func NewSpecialProjectService(repo repository.SpecialProjectRepository) *SpecialProjectService {
	return &SpecialProjectService{repo: repo}
}

// Converters
func toDomain(db *dbmodels.SpecialProjectDB) *models.SpecialProject {
	if db == nil {
		return nil
	}
	return &models.SpecialProject{
		ID:            db.ID,
		Title:         db.Title,
		Description:   db.Description,
		Image:         db.Image,
		IsActiveInBot: db.IsActiveInBot,
		CreatedAt:     db.CreatedAt,
		UpdatedAt:     db.UpdatedAt,
	}
}

func fromDomain(domain *models.SpecialProject) *dbmodels.SpecialProjectDB {

	if domain == nil {
		return nil
	}

	return &dbmodels.SpecialProjectDB{
		ID:            domain.ID,
		Title:         domain.Title,
		Description:   domain.Description,
		Image:         domain.Image,
		IsActiveInBot: domain.IsActiveInBot,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (s *SpecialProjectService) Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProject, error) {
	if proj.Title == "" {
		return nil, errors.New("title is required")
	}

	dbModel := fromDomain(proj)
	dbModel, err := s.repo.Create(ctx, dbModel)
	if err != nil {
		return nil, err
	}

	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) GetByID(ctx context.Context, id int64) (*models.SpecialProject, error) {
	dbModel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, dbmodels.ErrSpecProjNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toDomain(dbModel), nil
}

func (s *SpecialProjectService) List(ctx context.Context, statusStr string, search string) ([]models.SpecialProject, error) {
	var statusFilter *bool
	if statusStr != "" {
		val := statusStr == "active"
		statusFilter = &val
	}

	dbList, err := s.repo.List(ctx, statusFilter, search)
	if err != nil {
		return nil, err
	}

	result := make([]models.SpecialProject, 0, len(dbList))
	for _, item := range dbList {
		result = append(result, models.SpecialProject{
			ID:            item.ID,
			Title:         item.Title,
			IsActiveInBot: item.IsActiveInBot,
		})
	}

	return result, nil
}
