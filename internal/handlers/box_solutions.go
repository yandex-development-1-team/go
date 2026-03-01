package handlers

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/bot"
	"github.com/yandex-development-1-team/go/internal/logger"
	dbmodels "github.com/yandex-development-1-team/go/internal/repository/models"
	"github.com/yandex-development-1-team/go/internal/service"
	service_models "github.com/yandex-development-1-team/go/internal/service/models"
	"go.uber.org/zap"
	"time"
)

const (
	TextForBoxSolutions              = "📦 Коробочные решения\n\nВыберите интересующее вас предложение:\n"
	BoxSolutionsButtonBackToMainMenu = "back_to_main"
)

type BoxSolutionsRepository interface {
	GetBoxSolutions(ctx context.Context) ([]dbmodels.BoxSolution, error)
}

type BoxSolutionsHandler struct {
	bot     *bot.TelegramBot
	service *service.BoxSolutionsService
}

func NewBoxSolutions(bot *bot.TelegramBot, bsService *service.BoxSolutionsService) BoxSolutionsHandler {
	return BoxSolutionsHandler{
		bot:     bot,
		service: bsService,
	}
}

func (h *BoxSolutionsHandler) HandleBoxSolutions(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	ctxBoxSolutions, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger.Info("button is pressed",
		zap.String("user_id", query.Message.From.UserName),
		zap.String("service", query.Data),
	)

	boxSolutionsButtons, err := h.service.GetBoxSolutions(ctxBoxSolutions)
	if err != nil {
		logger.Error("failed to get inline buttons from service", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
	}

	reply := tgbotapi.NewMessage(query.Message.Chat.ID, TextForBoxSolutions)
	reply.ReplyMarkup = getMenuBoxSolutions(boxSolutionsButtons)

	if _, err := h.bot.Send(reply); err != nil {
		logger.Error("failed to send inline buttons for boxed solutions", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func getMenuBoxSolutions(boxSolutionsButtons []service_models.BoxSolutionsButton) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, boxSolution := range boxSolutionsButtons {
		btn := tgbotapi.NewInlineKeyboardButtonData(boxSolution.Name, boxSolution.Alias)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
