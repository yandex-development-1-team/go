package service

import (
	"context"
	"fmt"

	dbmodels "github.com/yandex-development-1-team/go/internal/database/repository/models"
	"github.com/yandex-development-1-team/go/internal/service/models"
)

type BoxSolutionsRepository interface {
	GetServices(ctx context.Context, telegramID int64) ([]dbmodels.Service, error)
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
	boxesDB, err := h.database.GetServices(ctx, telegramID)
	if err != nil {
		return []models.BoxSolutionsButton{}, fmt.Errorf("failed to retrieve boxed solutions from the database: %w", err)
	}

	boxSolutions := convertModelsDBToModels(boxesDB)
	boxSolutionsButtons := getMenuBoxSolutionsButtons(boxSolutions)

	return boxSolutionsButtons, nil
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

func convertModelsDBToModels(boxesDB []dbmodels.Service) []models.Service {
	var boxSolutions []models.Service

	for _, boxDB := range boxesDB {
		var availableSlots []models.AvailableSlot
		for _, availableSlotDB := range boxDB.AvailableSlots {
			availableSlot := models.AvailableSlot{
				Date:      availableSlotDB.Date,
				TimeSlots: availableSlotDB.TimeSlots,
			}

			availableSlots = append(availableSlots, availableSlot)
		}
		boxSolutions = append(boxSolutions, models.Service{
			ID:             boxDB.ID,
			Name:           boxDB.Name,
			Description:    boxDB.Description,
			AvailableSlots: availableSlots,
		})
	}

	return boxSolutions
}
