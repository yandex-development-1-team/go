package handlers

import (
	"context"
	"fmt"
	"github.com/yandex-development-1-team/go/internal/database/db_models"
	"github.com/yandex-development-1-team/go/internal/models"
)

type DataBaseClient interface {
	GetBoxSolutions(ctx context.Context) ([]db_models.BoxSolution, error)
}

type BoxSolutionsHandler struct {
	DBClient DataBaseClient
}

func NewBoxSolutions(dbClient DataBaseClient) BoxSolutionsHandler {
	return BoxSolutionsHandler{DBClient: dbClient}
}

//Логировать: user_id, выбранная услуга
func (bsh BoxSolutionsHandler) GetDetailsForBoxSolution(request models.GetDetailsForBoxSolutionRequest) {

}

func (bsh BoxSolutionsHandler) GetBoxSolutions(ctx context.Context) (models.BoxSolutionButtons, error) {
	//todo получены данные по боксам. В хендлере мы забираем только названия боксов. Где хранить оставшуюся информацию для быстрого доступа по кнопкам?
	boxesDB, err := bsh.DBClient.GetBoxSolutions(ctx)
	//todo обработку ошибки нужно обернуть во что-другое?
	if err != nil {
		return models.BoxSolutionButtons{}, fmt.Errorf("Ошибка получения коробочных решений: %w", err)
	}

	boxSolutions := convertModelsDBToModels(boxesDB)
	buttons := getButtons(boxSolutions)
	buttonsResp := models.BoxSolutionButtons{
		Description: "📦 Коробочные решения\n\nВыберите интересующее вас предложение:\n",
		Buttons:     buttons,
	}

	return buttonsResp, err
}

func convertModelsDBToModels(boxesDB []db_models.BoxSolution) models.GetBoxSolutionsResponse {
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
		box := models.BoxSolution{
			ID:             boxDB.ID,
			Name:           boxDB.Name,
			Description:    boxDB.Description,
			AvailableSlots: availableSlots,
		}

		response.Items = append(response.Items, box)
	}

	return response
}

func getButtons(response models.GetBoxSolutionsResponse) []models.Button {
	var buttonsResp []models.Button

	for _, boxSolution := range response.Items {
		buttonsResp = append(buttonsResp, models.Button{
			Alias: fmt.Sprintf("box_%d", boxSolution.ID),
			Name:  boxSolution.Name,
		})
	}

	buttonsResp = append(buttonsResp, models.Button{
		Alias: "back_to_main",
		Name:  "Назад",
	})

	return buttonsResp
}
