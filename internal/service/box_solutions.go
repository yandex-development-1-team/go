package service

import (
	"context"
	"fmt"
	dbmodels "github.com/yandex-development-1-team/go/internal/repository/models"
	"github.com/yandex-development-1-team/go/internal/service/models"
)

type BoxSolutionsRepository interface {
	GetBoxSolutions(ctx context.Context) ([]dbmodels.BoxSolution, error)
}

type BoxSolutionsService struct {
	database BoxSolutionsRepository
}

func NewBoxSolutionsService(repository BoxSolutionsRepository) *BoxSolutionsService {
	return &BoxSolutionsService{
		database: repository,
	}
}

func (h *BoxSolutionsService) GetBoxSolutions(ctx context.Context) ([]models.BoxSolutionsButton, error) {
	boxesDB, err := h.database.GetBoxSolutions(ctx)
	if err != nil {
		return []models.BoxSolutionsButton{}, fmt.Errorf("failed to retrieve boxed solutions from the database: %w", err)
	}

	boxSolutions := convertModelsDBToModels(boxesDB)
	boxSolutionsButtons := getMenuBoxSolutionsButtons(boxSolutions)

	return boxSolutionsButtons, nil
}

func getMenuBoxSolutionsButtons(boxSolutions []models.BoxSolution) []models.BoxSolutionsButton {
	var boxSolutionsButtons []models.BoxSolutionsButton

	for _, boxSolution := range boxSolutions {
		button := models.BoxSolutionsButton{
			Name:  boxSolution.Name,
			Alias: fmt.Sprintf("info:ID:%d", boxSolution.ID),
		}
		boxSolutionsButtons = append(boxSolutionsButtons, button)
	}

	btnBack := models.BoxSolutionsButton{
		Name:  "Назад",
		Alias: "back_to_main",
	}
	boxSolutionsButtons = append(boxSolutionsButtons, btnBack)

	return boxSolutionsButtons
}

func convertModelsDBToModels(boxesDB []dbmodels.BoxSolution) []models.BoxSolution {
	var boxSolutions []models.BoxSolution

	for _, boxDB := range boxesDB {
		var availableSlots []models.AvailableSlot
		for _, availableSlotDB := range boxDB.AvailableSlots {
			availableSlot := models.AvailableSlot{
				Date:      availableSlotDB.Date,
				TimeSlots: availableSlotDB.TimeSlots,
			}

			availableSlots = append(availableSlots, availableSlot)
		}
		boxSolutions = append(boxSolutions, models.BoxSolution{
			ID:             boxDB.ID,
			Name:           boxDB.Name,
			Description:    boxDB.Description,
			AvailableSlots: availableSlots,
		})
	}

	return boxSolutions
}
