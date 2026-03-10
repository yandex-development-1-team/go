package service

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

// BoxLister returns box solutions (services with box_solution=true) from storage.
type BoxLister interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
}

// APIBoxService implements HTTP API logic for boxed solutions.
type APIBoxService struct {
	lister BoxLister
}

// NewAPIBoxService creates a new instance of the box service.
func NewAPIBoxService(lister BoxLister) *APIBoxService {
	return &APIBoxService{lister: lister}
}

// List returns all box solutions for API (telegramID=0 — все коробки).
func (s *APIBoxService) List(ctx context.Context) ([]dto.BoxListItem, error) {
	services, err := s.lister.GetServices(ctx, 0)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BoxListItem, 0, len(services))
	for _, svc := range services {
		out = append(out, dto.BoxListItem{
			ID:          svc.ID,
			Name:        svc.Name,
			Description: svc.Description,
			Type:        svc.Type,
			BoxSolution: svc.BoxSolution,
		})
	}
	return out, nil
}
