package handlers

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
	"go.uber.org/zap"
)

const (
	TextForBoxSolutions = "📦 Коробочные решения\\n\\nВыберите интересующее вас предложение:\\n"
)

type DataBaseClient interface {
	GetBoxSolutions(ctx context.Context) ([]repository.BoxSolution, error)
}

type BoxSolutionsHandler struct {
	DBClient DataBaseClient
}

func NewBoxSolutions(dbClient DataBaseClient) BoxSolutionsHandler {
	return BoxSolutionsHandler{DBClient: dbClient}
}

func (h *Handler) HandleBoxSolutions(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	//todo нужно контекст создавать тут?
	//ctxBoxSolutions, cancel := context.WithTimeout(ctx, 2*time.Second)
	//defer cancel()

	logger.Info("button is pressed",
		zap.String("user_id", query.Message.From.UserName),
		zap.String("service", query.Data),
	)

	boxesDB, err := h.ClientBoxSolutions.GetBoxSolutions(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve boxed solutions from the database: %w", err)
	}

	boxSolutions := convertModelsDBToModels(boxesDB)
	buttons := getButtons(boxSolutions)
	//buttonsResp := models.BoxSolutionButtons{
	//	Description: TextForBoxSolutions,
	//	Buttons:     buttons,
	//}

	reply := tgbotapi.NewMessage(query.Message.Chat.ID, TextForBoxSolutions)
	reply.ReplyMarkup = menuBoxSolutions(buttons)

	if _, err := h.bot.Send(reply); err != nil {
		logger.Error("failed to send inline buttons for boxed solutions", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func menuBoxSolutions(buttons []models.Button) tgbotapi.InlineKeyboardMarkup {
	var rows []tgbotapi.InlineKeyboardButton
	for _, boxSolution := range buttons {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData(boxSolution.Name, boxSolution.Alias))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows)
}

func convertModelsDBToModels(boxesDB []repository.BoxSolution) models.GetBoxSolutionsResponse {
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
