package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"go.uber.org/zap"
)

// CallbackHandler интерфейс для обработчиков
type CallbackHandler interface {
	Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error
}

// CallbackRouter структура роутера
type CallbackRouter struct {
	handlers map[string]CallbackHandler
	bot      *tgbotapi.BotAPI
}

// NewCallbackRouter создает новый роутер
func NewCallbackRouter(bot *tgbotapi.BotAPI) *CallbackRouter {
	return &CallbackRouter{
		handlers: make(map[string]CallbackHandler),
		bot:      bot,
	}
}

// Register регистрирует обработчик
func (r *CallbackRouter) Register(name string, handler CallbackHandler) {
	r.handlers[name] = handler
	logger.Info("Callback handler registered", zap.String("name", name))
}

// HandleCallback обрабатывает callback запрос
func HandleCallback(router *CallbackRouter, query *tgbotapi.CallbackQuery) error {
	startTime := time.Now()

	logger.Info("CallbackQuery",
		zap.Int64("user_id", query.From.ID),
		zap.String("callback_data", query.Data),
		zap.Int("message_id", query.Message.MessageID),
	)

	defer func() {
		callbackConfig := tgbotapi.NewCallback(query.ID, "")
		if _, err := router.bot.Request(callbackConfig); err != nil {
			logger.Error("Failed to send answer", zap.Error(err),
				zap.String("callback_id", query.ID))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	handler, err := router.findHandler(query.Data)
	if err != nil {
		logger.Error("No handler found", zap.Error(err),
			zap.String("callback_data", query.Data))

		return err
	}

	if err := handler.Handle(ctx, query); err != nil {
		logger.Error("Handler execution failed", zap.Error(err),
			zap.String("handler_name", getName(query.Data)),
			zap.Duration("execution_time", time.Since(startTime)))

		return err
	}

	logger.Info("Callback processed successfully", zap.Int64("user_id", query.From.ID),
		zap.String("callback_data", query.Data), zap.Duration("processing_time", time.Since(startTime)),
	)

	return nil
}

// findHandler ищет обработчик данных
func (r *CallbackRouter) findHandler(data string) (CallbackHandler, error) {
	if handler, exists := r.handlers[data]; exists {
		return handler, nil
	}

	return nil, fmt.Errorf("Data handler not found")
}

// getName извлекает имя из данных callback
func getName(data string) string {
	parts := strings.Split(data, ":")

	if len(parts) > 0 {
		return parts[0]
	}

	return data
}
