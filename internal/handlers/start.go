package handlers

import (
	"context"
	"encoding/json"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/repository"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"

	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName string, firstName string, lastName string) error
}

type StartHandler struct {
	bot            BotAPI
	userRepository UserRepository
	session        repository.SessionRepository
}

func NewStartHandler(bot *tgbotapi.BotAPI, userRepository UserRepository, session repository.SessionRepository) *StartHandler {
	return &StartHandler{
		bot:            bot,
		userRepository: userRepository,
		session:        session,
	}
}

const (
	CallbackBoxSolutions    = "box_solutions"
	CallbackVisitGuide      = "visit_guide"
	CallbackSpecialProject  = "special_project"
	CallbackProjectExamples = "project_examples"
	CallbackAboutUs         = "about_us"
	CallbackSupport         = "support"
)

const (
	WelcomeText    = "👋 Добро пожаловать в Bot Яндекса!\n\nВыберите интересующую вас опцию:"
	ErrMessageUser = "Произошла ошибка, попробуйте позже."
)

func (sh *StartHandler) HandleStart(ctx context.Context, msg *tgbotapi.Message) error {
	metrics.IncMessagesReceived()

	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ObserveMessageProcessingDuration(duration)
	}()

	chatID := msg.Chat.ID
	telegramID := msg.From.ID
	username := msg.From.UserName
	firstName := msg.From.FirstName
	lastName := msg.From.LastName

	logger.Info("start command",
		zap.Int64("telegram_id", telegramID),
		zap.String("username", username),
		zap.Int64("chat_id", chatID),
	)

	state := sh.GetBookingState(ctx, telegramID)
	if state != nil && state.OldMessageID != nil {
		delTgMessage(sh.bot, &tgbotapi.Message{MessageID: *state.OldMessageID, Chat: &tgbotapi.Chat{ID: chatID}})
	}

	_ = sh.CreateSession(ctx, telegramID)

	if sh.userRepository != nil {
		if err := sh.userRepository.CreateUser(ctx, telegramID, username, firstName, lastName); err != nil {
			metrics.IncMessagesErrors()

			logger.Error("database error in CreateUser",
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
	}

	reply := tgbotapi.NewMessage(chatID, WelcomeText)
	reply.ReplyMarkup = mainMenuKeyboard()

	if _, err := sh.bot.Send(reply); err != nil {
		metrics.IncMessagesErrors()
		logger.Error("failed to send start message", zap.Int64("chat_id", chatID), zap.Error(err))
		return err
	}

	return nil
}

func (sh *StartHandler) Handle(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	delTgMessage(sh.bot, query.Message)

	reply := tgbotapi.NewMessage(query.Message.Chat.ID, WelcomeText)
	reply.ReplyMarkup = mainMenuKeyboard()

	if _, err := sh.bot.Send(reply); err != nil {
		logger.Error("failed to send start menu", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func mainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Коробочные решения", CallbackBoxSolutions),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Гайд по посещению", CallbackVisitGuide),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Запрос спецпроекта", CallbackSpecialProject),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Примеры спецпроектов", CallbackProjectExamples),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("О нас", CallbackAboutUs),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Связь с поддержкой", CallbackSupport),
		),
	)
}

// CreateSession creates the user session
func (sh *StartHandler) CreateSession(ctx context.Context, userID int64) error {
	if err := sh.session.ClearSession(ctx, userID); err != nil {
		logger.Error("failed to clear user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return err
	}

	data := map[string]interface{}{}
	err := sh.session.SaveSession(ctx, userID, "main_menu", data)
	if err != nil {
		logger.Error("failed to save user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return err
	}
	return nil
}

// GetBookingState returns the state of the booking process
func (sh *StartHandler) GetBookingState(ctx context.Context, userID int64) *botService.BookingState {
	session, err := sh.session.GetSession(ctx, userID)
	if err != nil {
		logger.Debug("Failed to get user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil
	}

	stateData, ok := session.StateData[botService.KeyForBookingData]
	if !ok {
		logger.Debug("state data not found in session")
		return nil
	}

	jsonData, err := json.Marshal(stateData)
	if err != nil {
		logger.Debug("failed to marshal state data", zap.Error(err))
		return nil
	}

	var state botService.BookingState
	if err := json.Unmarshal(jsonData, &state); err != nil {
		logger.Error("failed to unmarshal state data", zap.Error(err))
		return nil
	}

	return &state
}
