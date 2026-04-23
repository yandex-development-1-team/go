package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service"
)

const (
	textForBoxSolutions              = "📦 Коробочные решения\n\nВыберите интересующее вас предложение:\n"
	BoxSolutionsButtonBackToMainMenu = "main_menu"
	BoxSolutionsPerPage              = 5
)

type BoxSolutionsHandler struct {
	bot     *tgbotapi.BotAPI
	service *service.BoxSolutionsService
}

func NewBoxSolutions(bot *tgbotapi.BotAPI, bsService *service.BoxSolutionsService) *BoxSolutionsHandler {
	return &BoxSolutionsHandler{
		bot:     bot,
		service: bsService,
	}
}

func (h *BoxSolutionsHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	delTgMessage(h.bot, query.Message)

	ctxBoxSolutions, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	logger.Info("button is pressed",
		zap.String("user_id", query.Message.From.UserName),
		zap.String("service", query.Data),
	)

	page := parsePageFromCallback(query.Data)

	boxSolutionsButtons, err := h.service.GetBoxSolutions(ctxBoxSolutions, query.Message.Chat.ID)
	if err != nil {
		logger.Error("failed to get inline buttons from service", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
	}

	replyMarkup, pageText := getPaginatedBoxSolutionsMenu(boxSolutionsButtons, page)

	messageText := textForBoxSolutions
	if pageText != "" {
		messageText = pageText
	}

	reply := tgbotapi.NewMessage(query.Message.Chat.ID, messageText)
	reply.ReplyMarkup = replyMarkup

	if _, err := h.bot.Send(reply); err != nil {
		logger.Error("failed to send inline buttons for boxed solutions", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func getPaginatedBoxSolutionsMenu(boxSolutionsButtons []models.BoxSolutionsButton, currentPage int) (tgbotapi.InlineKeyboardMarkup, string) {
	totalItems := len(boxSolutionsButtons)
	totalPages := (totalItems + BoxSolutionsPerPage - 1) / BoxSolutionsPerPage

	if totalItems == 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		btnBack := getBackButton(BoxSolutionsButtonBackToMainMenu)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnBack))
		return tgbotapi.NewInlineKeyboardMarkup(rows...), "❌ Решения не найдены"
	}

	if currentPage < 1 {
		currentPage = 1
	}
	if currentPage > totalPages {
		currentPage = totalPages
	}

	startIdx := (currentPage - 1) * BoxSolutionsPerPage
	endIdx := startIdx + BoxSolutionsPerPage
	if endIdx > totalItems {
		endIdx = totalItems
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	for i := startIdx; i < endIdx; i++ {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			boxSolutionsButtons[i].Name,
			fmt.Sprintf("%s:%d", boxSolutionsButtons[i].Alias, currentPage),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	var paginationRow []tgbotapi.InlineKeyboardButton

	if currentPage > 1 {
		prevData := fmt.Sprintf("box_solutions:page:%d", currentPage-1)
		paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData("← Предыдущая", prevData))
	}

	if currentPage < totalPages {
		nextData := fmt.Sprintf("box_solutions:page:%d", currentPage+1)
		paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData("Следующая →", nextData))
	}

	if len(paginationRow) > 0 {
		rows = append(rows, paginationRow)
	}

	btnBack := getBackButton(BoxSolutionsButtonBackToMainMenu)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btnBack))

	var pageText string
	if totalPages > 1 {
		pageText = fmt.Sprintf("%s\n\n📄 Страница %d из %d", textForBoxSolutions, currentPage, totalPages)
	} else {
		pageText = textForBoxSolutions
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...), pageText
}

func parsePageFromCallback(data string) int {
	parts := strings.Split(data, ":")
	if len(parts) == 3 && parts[1] == "page" {
		page, _ := strconv.Atoi(parts[2])
		return page
	}
	return 1
}
