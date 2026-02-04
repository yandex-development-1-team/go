package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/bot"
	"go.uber.org/zap"
)

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
const WelcomeText = "👋 Добро пожаловать в Bot Яндекса!\n\nВыберите интересующую вас опцию:"

// UserSave сохраняет данные пользователя в БД
// Реализация может сохранять только новых пользователей
type UserSaver interface {
	SaveUser(userID int64, username string, chatID int64) error
}

var defaultUserSaver UserSaver

// SetUserSaver задает реализацию UserSaver
// Если не вызвана, сохранение пользователя не выполняется
func SetUserSaver(userSaver UserSaver) { defaultUserSaver = userSaver }

// HandleStart обрабатывает команду /start: логирует событие, при необходимости сохраняет
// пользователя через UserSaver, отправляет приветственное сообщение и главное меню с inline-кнопками
// Возвращает ошибку только при сбое отправки сообщения в Telegram
func HandleStart(bot *bot.TelegramBot, msg *tgbotapi.Message) error {
	userID := msg.From.ID
	chatID := msg.Chat.ID
	username := ""

	bot.Logger.Info("start command",
		zap.Int64("user_id", userID),
		zap.String("username", username),
		zap.Int64("chat_id", chatID),
	)

	if defaultUserSaver != nil {
		if err := defaultUserSaver.SaveUser(userID, username, chatID); err != nil {
			bot.Logger.Warn("failed to save user", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	keyboard := mainMenuKeyboard()
	reply := tgbotapi.NewMessage(chatID, WelcomeText)
	reply.ReplyMarkup = keyboard

	if _, err := bot.Api.Send(reply); err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

// mainMenuKeyboard возвращает разметку inline-кнопок главного меню
// Используется в HandleStart при отправке приветственного сообщения
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
