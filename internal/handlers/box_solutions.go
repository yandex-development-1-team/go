package handlers

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	dbmodels "github.com/yandex-development-1-team/go/internal/repository/models"
	"go.uber.org/zap"
	"time"
)

const (
	TextForBoxSolutions = "📦 Коробочные решения\n\nВыберите интересующее вас предложение:\n"
)

type BoxSolutionsRepository interface {
	GetBoxSolutions(ctx context.Context) ([]dbmodels.BoxSolution, error)
}

type BoxSolutionsHandlerBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type BoxSolutionsHandler struct {
	bot                    BoxSolutionsHandlerBot
	boxSolutionsRepository BoxSolutionsRepository
}

func NewBoxSolutions(bot BoxSolutionsHandlerBot, repository BoxSolutionsRepository) BoxSolutionsHandler {
	return BoxSolutionsHandler{
		bot:                    bot,
		boxSolutionsRepository: repository,
	}
}

func (h *BoxSolutionsHandler) HandleBoxSolutions(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	ctxBoxSolutions, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger.Info("button is pressed",
		zap.String("user_id", query.Message.From.UserName),
		zap.String("service", query.Data),
	)

	boxesDB, err := h.boxSolutionsRepository.GetBoxSolutions(ctxBoxSolutions)
	if err != nil {
		return fmt.Errorf("failed to retrieve boxed solutions from the database: %w", err)
	}

	boxSolutions := convertModelsDBToModels(boxesDB)

	reply := tgbotapi.NewMessage(query.Message.Chat.ID, TextForBoxSolutions)
	reply.ReplyMarkup = getMenuBoxSolutions(boxSolutions)

	if _, err := h.bot.Send(reply); err != nil {
		logger.Error("failed to send inline buttons for boxed solutions", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func getMenuBoxSolutions(boxSolutions []models.BoxSolution) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, boxSolution := range boxSolutions {
		alias := fmt.Sprintf("info:ID:%d", boxSolution.ID)
		btn := tgbotapi.NewInlineKeyboardButtonData(boxSolution.Name, alias)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	btn := tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_main")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
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
