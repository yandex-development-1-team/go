package handlers

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

const cmdStart = "start"

type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	// новые методы для bot api добавлять сюда, а реализовывать в go/internal/bot/bot.go
}

type Handler struct {
	startHandler        StartHandler
	boxSolutionsHandler BoxSolutionsHandler
}

func NewHandler(startHandler StartHandler, boxSolutionsHandler BoxSolutionsHandler) *Handler {
	return &Handler{
		startHandler:        startHandler,
		boxSolutionsHandler: boxSolutionsHandler,
	}
}

func (h *Handler) Handle(ctx context.Context, update tgbotapi.Update) {
	if msg := update.Message; msg != nil {
		if msg.IsCommand() {
			switch msg.Command() {
			case cmdStart:
				if err := h.startHandler.HandleStart(msg); err != nil {
					logger.Error("failed to handle /start", zap.Error(err))
				}
			// в новые ветки добавлять вызовы функций обработчиков команд
			default:
				// todo
			}
		}
		// todo
	}
	if callbackQuery := update.CallbackQuery; callbackQuery != nil {
		switch callbackQuery.Data {
		case CallbackBoxSolutions:
			if err := h.boxSolutionsHandler.HandleBoxSolutions(ctx, callbackQuery); err != nil {
				logger.Error("failed to handle callback BoxSolutions", zap.Error(err))
			}
		}
		//todo
	}
}
