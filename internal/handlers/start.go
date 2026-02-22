package handlers

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"

	"go.uber.org/zap"
)

type UserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName string, firstName string, lastName string) error
}

type StartHandlerBot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type StartHandler struct {
	bot            StartHandlerBot
	userRepository UserRepository
}

func NewStartHandler(bot StartHandlerBot, userRepository UserRepository) StartHandler {
	return StartHandler{
		bot:            bot,
		userRepository: userRepository,
	}
}

// Callback-данные для inline-кнопок главного меню
// Используются как callback_data при нажатии на кнопки после /start
const (
	CallbackBoxSolutions    = "box_solutions"
	CallbackVisitGuide      = "visit_guide"
	CallbackSpecialProject  = "special_project"
	CallbackProjectExamples = "project_examples"
	CallbackAboutUs         = "about_us"
	CallbackSupport         = "support"
)

// WelcomeText - Приветственное сообщение при команде /start
// ErrMessageUser - Текст об ошибке при работе с БД
const (
	WelcomeText    = "👋 Добро пожаловать в Bot Яндекса!\n\nВыберите интересующую вас опцию:"
	ErrMessageUser = "Произошла ошибка, попробуйте позже."
)

// HandleStart обрабатывает команду /start

func (sh *StartHandler) HandleStart(msg *tgbotapi.Message) error {

	// Инкрементируем MessagesReceived
	metrics.IncMessagesReceived()

	// Засекаем время для MessageProcessingDuration
	startTime := time.Now()

	// Записываем длительность обработки в конце
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ObserveMessageProcessingDuration(duration)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	chatID := msg.Chat.ID
	telegramID := msg.From.ID
	username := ""
	firstName := ""
	lastName := ""
	if msg.From != nil {
		username = msg.From.UserName
		firstName = msg.From.FirstName
		lastName = msg.From.LastName
	}

	logger.Info("start command",
		zap.Int64("telegram_id", telegramID),
		zap.String("username", username),
		zap.Int64("chat_id", chatID),
	)

	if sh.userRepository != nil {
		if err := sh.userRepository.CreateUser(
			ctx,
			telegramID,
			username,
			firstName,
			lastName,
		); err != nil {

			// При ошибке создания пользователя инкрементируем MessagesErrors
			metrics.IncMessagesErrors()

			logger.Error("database error in CreateUser",
				zap.Int64("telegram_id", telegramID),
				zap.String("username", username),
				zap.Error(err),
			)

			errMsg := tgbotapi.NewMessage(chatID, ErrMessageUser)
			if _, sendErr := sh.bot.Send(errMsg); sendErr != nil {
				logger.Error("failed to send error message", zap.Error(sendErr))

				// При ошибке отправки сообщения об ошибке инкрементируем MessagesErrors
				metrics.IncMessagesErrors()
			}

			return err
		}
	}

	reply := tgbotapi.NewMessage(chatID, WelcomeText)
	reply.ReplyMarkup = mainMenuKeyboard()

	if _, err := sh.bot.Send(reply); err != nil {

		// При ошибке отправки инкрементируем MessagesErrors
		metrics.IncMessagesErrors()
		logger.Error("failed to send start message", zap.Int64("chat_id", chatID), zap.Error(err))
		return err
	}

	return nil
}

func (sh *StartHandler) HandleStartBackToMainMenu(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	reply := tgbotapi.NewMessage(query.Message.Chat.ID, WelcomeText)
	reply.ReplyMarkup = mainMenuKeyboard()

	if _, err := sh.bot.Send(reply); err != nil {
		logger.Error("failed to send start menu for button BoxSolutionsButtonBackToMainMenu", zap.Int64("chat_id", query.Message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

// mainMenuKeyboard возвращает разметку inline-кнопок главного меню
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
