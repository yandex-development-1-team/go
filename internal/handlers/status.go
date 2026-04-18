package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/repository/postgres"
)

// StatusHandler processes a command '/status'
type StatusHandler struct {
	bot     BotAPI
	repo    *postgres.BookingRepo
	session repository.SessionRepository
}

// NewStatusHandler creates a new instance of the 'StatusHandler'
func NewStatusHandler(bot *tgbotapi.BotAPI, repo *postgres.BookingRepo, session repository.SessionRepository) *StatusHandler {
	return &StatusHandler{
		bot:     bot,
		repo:    repo,
		session: session,
	}
}

// Handle processes the '/status' command
func (sh *StatusHandler) Handle(ctx context.Context, msg *tgbotapi.Message) error {
	metrics.IncMessagesReceived()
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ObserveMessageProcessingDuration(duration)
	}()

	chatID := msg.Chat.ID
	telegramID := msg.From.ID
	username := msg.From.UserName

	logger.Info("status command",
		zap.Int64("telegram_id", telegramID),
		zap.String("username", username),
		zap.Int64("chat_id", chatID),
	)

	bookings, err := sh.repo.GetExtBookingsByUserID(ctx, telegramID)
	if err != nil {
		metrics.IncMessagesErrors()

		logger.Error("database error in GetBookingsByUserID",
			zap.Int64("telegram_id", telegramID),
			zap.String("username", username),
			zap.Error(err),
		)

		errMsg := tgbotapi.NewMessage(chatID, ErrMessageUser)
		if _, sendErr := sh.bot.Send(errMsg); sendErr != nil {
			logger.Error("failed to send error message", zap.Error(sendErr))
			metrics.IncMessagesErrors()
		}
		return err
	}

	text := sh.formatBookingMessage(bookings)

	reply := tgbotapi.NewMessage(chatID, text)

	if _, err := sh.bot.Send(reply); err != nil {
		metrics.IncMessagesErrors()
		logger.Error("failed to send start message", zap.Int64("chat_id", chatID), zap.Error(err))
		return err
	}

	return nil
}

// formatBookingMessage formats booking messages
func (sh *StatusHandler) formatBookingMessage(bookings []models.BookingExt) string {
	if len(bookings) == 0 {
		return "У вас пока нет бронирований."
	}

	var message strings.Builder

	for i, booking := range bookings {
		bookingDate := booking.BookingDate.Format("02.01.2006")

		var bookingTimeStr string
		if booking.BookingTime != nil {
			bookingTimeStr = booking.BookingTime.Format("15:04")
		} else {
			bookingTimeStr = "Не указано"
		}

		fmt.Fprintf(&message, "%d. %s\n"+
			"   Дата: %s\n"+
			"   Время: %s\n"+
			"   Гость: %s\n"+
			"   Статус: %s\n"+
			"   ID: %d\n\n",
			i+1,
			booking.ServiceName,
			bookingDate,
			bookingTimeStr,
			booking.GuestName,
			booking.Status,
			booking.ID)
	}

	return message.String()
}
