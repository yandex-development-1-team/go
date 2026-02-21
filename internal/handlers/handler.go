package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"go.uber.org/zap"
)

const cmdStart = "start"

type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	// новые методы для bot api добавлять сюда, а реализовывать в go/internal/bot/bot.go
}

type Handler struct {
	bot                 Bot
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

	// Нужно получить количество активных пользователей
	activeUsers := getActiveUsersCount(ctx) // Эту функцию нужно реализовать

	// И далее обновляем ActiveUsers перед обработкой
	metrics.SetActiveUsers(activeUsers)

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
		case BoxSolutionsButtonBackToMainMenu:
			if err := h.startHandler.HandleStartBackToMainMenu(ctx, callbackQuery); err != nil {
				logger.Error("failed to handle callback BoxSolutionsButtonBackToMainMenu", zap.Error(err))
			}
			//todo
		}
	}
}

// Заготовка для реализации функции
func getActiveUsersCount(ctx context.Context) int {
	// TODO: заменить на реальный подсчет активных пользователей
	// Например:
	// count, err := userRepo.GetActiveUsersCount(ctx)
	// if err != nil {
	//     logger.Error("failed to get active users", zap.Error(err))
	//     return 0
	// }
	// return count

	// Чтобы тесты проходили, пока можно возвращать 1
	return 1
}
