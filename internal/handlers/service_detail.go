package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// ServiceType определяет тип услуги для выбора правильного набора кнопок
type ServiceType string

const (
	ServiceTypeMuseum  ServiceType = "museum"
	ServiceTypeSport   ServiceType = "sport"
	ServiceTypeDefault ServiceType = "default"
)

// internal/handlers/service_detail.go
func (h *ServiceHandler) HandleServiceDetail(ctx context.Context, tg *tgbotapi.CallbackQuery) error {
	// ШАГ 1: Получаем ID пользователя и номер чата
	userID := tg.Message.UserID
	serviceID := tg.Message.ChatID
	// ШАГ 2: Логирование запроса
	h.logger.Info("service_detail_requested",
		zap.Int("service_id", serviceID),
		zap.Int64("user_id", userID),
	)

	// ШАГ 3: Получение данных из репозитория
	service, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return h.handleError(userID, serviceID, err)
	}
	serviceName := service.Name
	if serviceName == "" {
		serviceName = "Прочее"
	}

	// ШАГ 4: Формирование сообщения (чистая функция)
	messageText := buildServiceMessage(service, serviceName)

	// ШАГ 5: Генерация клавиатуры (сервис)
	keyboard := h.keyboardService.ServiceDetailKeyboard(
		service.Type,
		service.ID,
		service.BoxID,
	)

	// ШАГ 6: Отправка сообщения
	if err := h.sendMessage(userID, messageText, keyboard); err != nil {
		h.logger.Error("failed_to_send_service_detail",
			zap.Int("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send service detail: %w", err)
	}

	// ШАГ 7: Логирование успешного показа
	h.logger.Info("service_detail_shown",
		zap.Int("service_id", serviceID),
		zap.String("service_name", service.Name),
		zap.Int64("user_id", userID),
		zap.String("service_type", string(service.Type)),
	)

	return nil
}

// Вспомогательные методы для оркестратора
func (h *ServiceHandler) handleError(userID int64, serviceID int, err error) error {
	// Отправка сообщения об ошибке + логирование
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
	}
	return err
}

func buildServiceMessage(service *Service, serviceName string) string {
	var builder strings.Builder

	// Эмодзи в зависимости от типа услуги
	emoji := "✨"
	switch service.Type {
	case ServiceTypeMuseum:
		emoji = "🎨"
	case ServiceTypeSport:
		emoji = "⚽"
	}

	builder.WriteString(emoji)
	builder.WriteString(" ")
	builder.WriteString(serviceName)
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
		}
		if i == len(sections)-1 {
			builder.WriteString("\n")
		}
	}

	// Призыв к действию
	switch service.Type {
	case ServiceTypeMuseum:
		builder.WriteString("Выберите тип посещения:")
	case ServiceTypeSport:
		builder.WriteString("Выберите действие:")
	default:
		builder.WriteString("Доступные действия:")
	}

	return builder.String()
}
