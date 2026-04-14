package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// RequestSpHandler processes the guide's request
type RequestSpHandler struct {
	service  *botService.RequestSpService
	bot      *tgbotapi.BotAPI
	keyboard *KeyboardService
	sh       *StartHandler
	bs       *BoxSolutionsHandler
}

// NewRequestSpHandler creates a new instance of the 'RequestSpHandler'
func NewRequestSpHandler(service *botService.RequestSpService, bot *tgbotapi.BotAPI, sh *StartHandler, bs *BoxSolutionsHandler, keyboard *KeyboardService) *RequestSpHandler {
	return &RequestSpHandler{
		service:  service,
		bot:      bot,
		keyboard: keyboard,
		sh:       sh,
		bs:       bs,
	}
}

// Handle processes the 'request sp' button
func (h *RequestSpHandler) Handle(ctx context.Context, tg *tgbotapi.CallbackQuery) error {
	userID := tg.From.ID
	chatID := tg.Message.Chat.ID

	logger.Info("'request sp' requested", zap.Int64("user_id", userID))

	res, err := h.service.GetBySlug(ctx)
	if err != nil {
		return h.handleError(chatID, userID, err)
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "*%s*\n\n", res.Title)
	builder.WriteString(res.Content)
	builder.WriteString("\n\n")

	if len(res.Links) > 0 {
		builder.WriteString("*Ссылки:*\n")
		for i, link := range res.Links {
			title := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, link.Title)
			url := link.URL
			fmt.Fprintf(&builder, "%d. [%s](%s)\n", i+1, title, url)
		}
	}

	msg := tgbotapi.NewMessage(chatID, builder.String())
	msg.ParseMode = "Markdown"
	keyboard := h.keyboard.CreateButton("Назад", "main_menu")
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// sendError sends an error message
func (h *RequestSpHandler) sendError(chatID int64, errorMsg string) error {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка: %s", errorMsg))
	_, err := h.bot.Send(msg)
	return err
}

// handleError logs failures and sends a message
func (h *RequestSpHandler) handleError(chatID int64, userID int64, err error) error {
	logger.Error("failed to get req_sp",
		zap.Int64("user_id", userID),
		zap.Error(err),
	)

	if sendErr := h.sendError(chatID, "Не удалось сделать запрос спецпроекта"); sendErr != nil {
		return sendErr
	}

	return err
}
