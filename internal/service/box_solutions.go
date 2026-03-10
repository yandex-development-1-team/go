package service

import (
	"context"
	"fmt"

	"github.com/yandex-development-1-team/go/internal/models"
)

type BoxSolutionsRepository interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
}

type BoxSolutionsService struct {
	database BoxSolutionsRepository
}

func NewBoxSolutionsService(repository BoxSolutionsRepository) *BoxSolutionsService {
	return &BoxSolutionsService{
		database: repository,
	}
}

func (h *BoxSolutionsService) GetBoxSolutions(ctx context.Context, telegramID int64) ([]models.BoxSolutionsButton, error) {
	boxSolutions, err := h.database.GetServices(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	return getMenuBoxSolutionsButtons(boxSolutions), nil
}

func getMenuBoxSolutionsButtons(boxSolutions []models.Service) []models.BoxSolutionsButton {
	var boxSolutionsButtons []models.BoxSolutionsButton

	for _, boxSolution := range boxSolutions {
		button := models.BoxSolutionsButton{
			Name:  boxSolution.Name,
			Alias: fmt.Sprintf("info:ID:%d", boxSolution.ID),
		}
		boxSolutionsButtons = append(boxSolutionsButtons, button)
	}

	return boxSolutionsButtons
}
