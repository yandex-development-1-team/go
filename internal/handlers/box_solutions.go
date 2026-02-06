package handlers

import (
	"context"
	"github.com/yandex-development-1-team/go/internal/database/db_models"
	"github.com/yandex-development-1-team/go/internal/models"
)

type DataBaseClient interface {
	GetBoxes(ctx context.Context) ([]db_models.Box, error)
}

type BoxSolutionsHandler struct {
	DBClient DataBaseClient
}

func NewBoxSolutions(dbClient DataBaseClient) BoxSolutionsHandler {
	return BoxSolutionsHandler{DBClient: dbClient}
}

func (bsh BoxSolutionsHandler) GetBoxSolutions(ctx context.Context) (models.GetBoxSolutionsResponse, error) {
	boxesDB, err := bsh.DBClient.GetBoxes(ctx)
	//todo добавить обработку ошибки

	boxSolutionsResponse := convertModelsDBToModels(boxesDB)

	return boxSolutionsResponse, err
}

func convertModelsDBToModels(boxesDB []db_models.Box) models.GetBoxSolutionsResponse {
	var response models.GetBoxSolutionsResponse

	for _, boxDB := range boxesDB {
		var availableSlots []models.AvailableSlot
		for _, availableSlotDB := range boxDB.AvailableSlots {
			availableSlot := models.AvailableSlot{
				Date:      availableSlotDB.Date,
				TimeSlots: availableSlotDB.TimeSlots,
			}

			availableSlots = append(availableSlots, availableSlot)
		}
		box := models.Box{
			Name:           boxDB.Name,
			Description:    boxDB.Description,
			AvailableSlots: availableSlots,
		}

		response.Items = append(response.Items, box)
	}

	return response
}
