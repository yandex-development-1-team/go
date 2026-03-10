package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/yandex-development-1-team/go/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
)

const VessagesErrData = "❌ К сожалению, не удалось загрузить информацию об услуге. Пожалуйста, попробуйте позже."

// internal/handlers/service_detail.go
func (h *ServiceHandler) HandleServiceDetail(ctx context.Context, tg *tgbotapi.CallbackQuery) error {
	// ШАГ 1: Получаем ID пользователя и номер чата
	userID := tg.From.ID
	chatID := tg.Message.Chat.ID
	callbackData := tg.Data
	serviceID, err := parseServiceID(callbackData)
	if err != nil {
		logger.Error("failed_to_parse_service_id",
			zap.String("callback_data", callbackData),
			zap.Int64("user_id", userID),
			zap.Int64("chat_id", chatID),
			zap.Error(err),
		)
		// Отправляем сообщение об ошибке В ЧАТ
		errorMsg := tgbotapi.NewMessage(chatID, VessagesErrData)
		errorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, sendErr := h.bot.Send(errorMsg); sendErr != nil {
			logger.Error("failed_to_send_error_message", zap.Error(sendErr))
		}
		return err
	}
	// ШАГ 2: Логирование запроса
	logger.Info("service_detail_requested",
		zap.Int("service_id", serviceID),
		zap.Int64("user_id", userID),
	)

	// ШАГ 3: Получение данных из репозитория
	serviceDBModel, err := h.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return h.handleError(userID, serviceID, err)
	}

	service := convertDBServiceModelToHandlerModel(serviceDBModel)
	serviceName := service.Name
	if serviceName == "" {
		serviceName = "Прочее"
	}

	// ШАГ 4: Формирование сообщения (чистая функция)
	messageText := buildServiceMessage(&service, serviceName)

	// ШАГ 5: Генерация клавиатуры (сервис)
	keyboard := h.keyboardService.ServiceDetailKeyboard(
		service.Type,
		service.ID,
	)

	// ШАГ 6: Отправка сообщения
	if err := h.sendMessage(chatID, messageText, keyboard); err != nil {
		logger.Error("failed_to_send_service_detail",
			zap.Int("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send service detail: %w", err)
	}

	// ШАГ 7: Логирование успешного показа
	logger.Info("service_detail_shown",
		zap.Int("service_id", serviceID),
		// zap.String("service_name", service.Name),
		zap.Int64("user_id", userID),
		// zap.String("service_type", string(service.Type)),
	)

	return nil
}

func convertDBServiceModelToHandlerModel(dbServiceModel models.Service) Service {
	slots := make([]AvailableSlot, 0, len(dbServiceModel.AvailableSlots))
	for _, s := range dbServiceModel.AvailableSlots {
		slots = append(slots, AvailableSlot{
			Date:      s.Date,
			TimeSlots: s.TimeSlots,
		})
	}

	return Service{
		ID:             dbServiceModel.ID,
		Name:           dbServiceModel.Name,
		Description:    dbServiceModel.Description,
		Rules:          dbServiceModel.Rules,
		Schedule:       dbServiceModel.Schedule,
		AvailableSlots: slots,
		Type:           ServiceType(dbServiceModel.Type),
	}
}

// Вспомогательные методы для оркестратора
func (h *ServiceHandler) handleError(userID int64, serviceID int, err error) error {
	// Отправка сообщения об ошибке + логирование
	if err != nil {
		logger.Error("failed_to_get_service",
			zap.Int("service_id", serviceID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)

		// Отправляем сообщение об ошибке пользователю
		msg := tgbotapi.NewMessage(userID, "❌ К сожалению, не удалось загрузить информацию об услуге. Пожалуйста, попробуйте позже.")
		_, sendErr := h.bot.Send(msg)
		if sendErr != nil {
			logger.Error("failed_to_send_error_message",
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
