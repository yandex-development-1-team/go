package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/repository" // название пакета с запросом к ДБ
)

// ServiceHandler обрабатывает действия, связанные с услугами
type ServiceHandler struct {
	logger *zap.Logger
	repo   *repository.Repository
	bot    *tgbotapi.BotAPI
}

// ServiceType определяет тип услуги для выбора правильного набора кнопок
type ServiceType string

const (
	ServiceTypeMuseum  ServiceType = "museum"
	ServiceTypeSport   ServiceType = "sport"
	ServiceTypeDefault ServiceType = "default"
)

// Service представляет собой услугу с полной информацией
type Service struct {
	ID          int         //Уникальный идентификатор услуги
	Name        string      // название услуги
	Description string      // описание
	Rules       string      // правила
	Schedule    string      // время проведения
	Type        ServiceType // Тип услуги (музей, спорт и т.д.)
	BoxID       int         // ID бокса/категории для кнопки "Назад"
}

// NewServiceHandler создаёт новый обработчик услуг
func NewServiceHandler(logger *zap.Logger, repo *repository.Repository, bot *tgbotapi.BotAPI) *ServiceHandler {
	return &ServiceHandler{
		logger: logger.Named("service_handler"),
		repo:   repo,
		bot:    bot,
	}
}

// HandleServiceDetail обрабатывает нажатие на конкретную услугу
func (h *ServiceHandler) HandleServiceDetail(ctx context.Context, serviceID int, userID int64) error {

	// === ШАГ 1: Логирование запроса ===
	h.logger.Info("service_detail_requested",
		zap.Int("service_id", serviceID),
		zap.Int64("user_id", userID),
	)

	// === ШАГ 2: Получение информации об услуге из базы данных ===
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		h.logger.Error("failed_to_get_service",
			zap.Int("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)

		// Отправляем сообщение об ошибке пользователю
		msg := tgbotapi.NewMessage(userID, "❌ К сожалению, не удалось загрузить информацию об услуге. Пожалуйста, попробуйте позже.")
		_, sendErr := h.bot.Send(msg)
		if sendErr != nil {
			h.logger.Error("failed_to_send_error_message",
				zap.Int64("user_id", userID),
				zap.Error(sendErr),
			)
		}

		return err
	}

	// === ШАГ 3: Формирование сообщения с информацией об услуге ===
	// Используем эмодзи в зависимости от типа услуги для визуального выделения
	emoji := "✨" // эмодзи по дефолту
	switch service.Type {
	case ServiceTypeMuseum:
		emoji = "🎨"
	case ServiceTypeSport:
		emoji = "⚽"
	}

	var builder strings.Builder
	builder.Grow(300)

	// Заголовок с эмодзи
	builder.WriteString(emoji)
	builder.WriteString(" ")
	builder.WriteString(service.Name)
	builder.WriteString("\n\n")

	sections := []struct {
		label string
		value string
	}{
		{"Описание", service.Description},
		{"Правила", service.Rules},
		{"Расписание", service.Schedule},
	}

	for i, section := range sections {
		if section.value != "" {
			builder.WriteString(section.label)
			builder.WriteString(": ")
			builder.WriteString(section.value)
			builder.WriteString("\n")

			// Добавляем перенос только если это НЕ последний непустой раздел
			// или если после него будут кнопки

		}
		if i == len(sections)-1 {
			builder.WriteString("\n")
		}
	}

	// Формируем основное сообщение

	// === ШАГ 4: Формирование призыва к действию и кнопок ===
	// Добавляем разделитель и призыв к действию

	switch service.Type {
	case ServiceTypeMuseum:
		builder.WriteString("Выберите тип посещения:")
	case ServiceTypeSport:
		builder.WriteString("Выберите действие:")
	default:
		builder.WriteString("Доступные действия:")
	}

	messageText := builder.String() //итоговая строка

	// === ШАГ 5: Создание клавиатуры с кнопками ===
	// Кнопки зависят от типа услуги
	var buttons [][]tgbotapi.InlineKeyboardButton

	switch service.Type {
	case ServiceTypeMuseum:
		// Кнопки для музеев/галерей
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("👤 Приватный тур", fmt.Sprintf("private_view_%d", service.ID)),
				tgbotapi.NewInlineKeyboardButtonData("👥 Групповой тур", fmt.Sprintf("public_view_%d", service.ID)),
			},
		}
	case ServiceTypeSport:
		// Кнопка для спортивных услуг
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать сейчас", fmt.Sprintf("book_now_%d", service.ID)),
			},
		}
	default:
		// Универсальная кнопка бронирования для других типов
		buttons = [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("📅 Забронировать", fmt.Sprintf("book_now_%d", service.ID)),
			},
		}
	}

	// Всегда добавляем кнопку "Назад" в отдельную строку
	// Возвращаемся к списку услуг в том же боксе/категории
	backButton := tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("back_to_box_%d", service.BoxID))
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	// Создаём разметку клавиатуры
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// === ШАГ 6: Отправка сообщения пользователю ===
	msg := tgbotapi.NewMessage(userID, messageText)
	msg.ReplyMarkup = keyboard

	_, err = h.bot.Send(msg)
	if err != nil {
		h.logger.Error("failed_to_send_service_detail",
			zap.Int("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send service detail message: %w", err)
	}

	// === ШАГ 7: Логирование успешного показа ===
	h.logger.Info("service_detail_shown",
		zap.Int("service_id", serviceID),
		zap.String("service_name", service.Name),
		zap.Int64("user_id", userID),
		zap.String("service_type", string(service.Type)),
	)

	return nil
}
