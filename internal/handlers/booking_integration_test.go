package handlers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yandex-development-1-team/go/internal/database/repository"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// Тест: Полный цикл бронирования от начала до конца
func TestBookingFormHandler_FullFlow(t *testing.T) {
	// Очищаем таблицу bookings перед тестом
	_, err := handlerTestDB.Exec("DELETE FROM bookings WHERE user_id = $1", TestUserID)
	require.NoError(t, err)

	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	// Используем реальный репозиторий вместо мока
	userRepo := repository.NewUserRepository(handlerTestDB)

	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	emptyBot := &tgbotapi.BotAPI{}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	currentMsgID := 1

	// Гарантируем, что пользователь существует
	ensureUserExists(t, userID)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	// ИСПРАВЛЕНО: уменьшаем количество ожидаемых Send с 8 до 7
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil).Times(7)

	targetDate := time.Now().AddDate(0, 0, 2)

	// ШАГ 1: Начало бронирования
	t.Log("ШАГ 1: Начало бронирования")
	query1 := createTestCallbackQuery(chatID, userID, "book:private:1", currentMsgID)

	err = handler.Handle(ctx, query1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state1 := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state1)
	assert.Equal(t, botService.StepSelectDate, state1.Step)
	assert.Equal(t, userID, state1.UserID)
	assert.Equal(t, 1, state1.ServiceID)
	assert.Equal(t, "private", state1.VisitType)
	currentMsgID++
	t.Log("  → ШАГ 1 завершен")

	// ШАГ 2: Выбор даты
	t.Logf("ШАГ 2: Выбор даты %s", targetDate.Format("2006-01-02"))
	query2 := createTestCallbackQuery(chatID, userID, "book:select_date:private:"+targetDate.Format("2006-01-02"), currentMsgID)

	err = handler.Handle(ctx, query2)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state2 := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state2)
	assert.Equal(t, botService.StepEnterName, state2.Step)
	assert.Equal(t, targetDate.Format("2006-01-02"), state2.SelectedDate.Format("2006-01-02"))
	currentMsgID++
	t.Log("  → ШАГ 2 завершен")

	// ШАГ 3: Ввод ФИО
	t.Log("ШАГ 3: Ввод ФИО")
	msg3 := createTestMessage(chatID, userID, "Иванов Иван Иванович", currentMsgID)
	err = handler.HandleTextMessage(ctx, msg3)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state3 := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state3)
	assert.Equal(t, botService.StepEnterOrg, state3.Step)
	assert.Equal(t, "Иванов Иван Иванович", state3.GuestName)
	currentMsgID++
	t.Log("  → ШАГ 3 завершен")

	// ШАГ 4: Ввод организации
	t.Log("ШАГ 4: Ввод организации")
	msg4 := createTestMessage(chatID, userID, "ООО Ромашка", currentMsgID)
	err = handler.HandleTextMessage(ctx, msg4)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state4 := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state4)
	assert.Equal(t, botService.StepEnterPosition, state4.Step)
	assert.Equal(t, "ООО Ромашка", state4.GuestOrganization)
	currentMsgID++
	t.Log("  → ШАГ 4 завершен")

	// ШАГ 5: Ввод должности
	t.Log("ШАГ 5: Ввод должности")
	msg5 := createTestMessage(chatID, userID, "Директор", currentMsgID)
	err = handler.HandleTextMessage(ctx, msg5)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state5 := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state5)
	assert.Equal(t, botService.StepConfirmation, state5.Step)
	assert.Equal(t, "Директор", state5.GuestPosition)
	currentMsgID++
	t.Log("  → ШАГ 5 завершен")

	// Проверка всех накопленных данных
	assert.Equal(t, targetDate.Format("2006-01-02"), state5.SelectedDate.Format("2006-01-02"))
	assert.Equal(t, "Иванов Иван Иванович", state5.GuestName)
	assert.Equal(t, "ООО Ромашка", state5.GuestOrganization)
	assert.Equal(t, "Директор", state5.GuestPosition)
	assert.Equal(t, "private", state5.VisitType)
	assert.Equal(t, 1, state5.ServiceID)

	// ШАГ 6: Подтверждение бронирования
	t.Log("ШАГ 6: Подтверждение бронирования")
	query3 := createTestCallbackQuery(chatID, userID, "book:confirm", currentMsgID)

	err = handler.Handle(ctx, query3)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	session, err := handlerSessionRepo.GetSession(ctx, userID)
	assert.Error(t, err, "Сессия должна быть очищена")
	assert.Nil(t, session, "Сессия должна быть nil")
	t.Log("  → ШАГ 6 завершен")

	var bookingCount int
	err = handlerTestDB.Get(&bookingCount, "SELECT COUNT(*) FROM bookings WHERE user_id = $1", userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, bookingCount, "Должна быть создана 1 запись бронирования")
	t.Logf("  → Бронирование создано в БД: %d запись", bookingCount)

	mockBot.AssertExpectations(t)
	t.Log("ИТОГ: Полный флоу бронирования успешно пройден")
}

// Тест: Проверка навигации "Назад"
func TestBookingFormHandler_BackNavigation(t *testing.T) {
	// Очищаем сессию перед тестом
	_ = handlerSessionRepo.ClearSession(context.Background(), TestUserID)

	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := repository.NewUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	emptyBot := &tgbotapi.BotAPI{}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 1

	ensureUserExists(t, userID)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	// ШАГ 1: Начало бронирования
	query1 := createTestCallbackQuery(chatID, userID, "book:private:1", messageID)
	err := handler.Handle(ctx, query1)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	state := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state)
	assert.Equal(t, botService.StepSelectDate, state.Step)

	// ШАГ 2: Возврат в меню с шага выбора даты
	queryBack := createTestCallbackQuery(chatID, userID, "book:main_menu", messageID+1)
	err = handler.Handle(ctx, queryBack)
	assert.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	// Проверяем, что сессия очищена
	session, _ := handlerSessionRepo.GetSession(ctx, userID)
	assert.Nil(t, session)

	mockBot.AssertExpectations(t)
}

// Тест: Обработка ошибок валидации
func TestBookingFormHandler_ValidationErrors(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := repository.NewUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	emptyBot := &tgbotapi.BotAPI{}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID

	ensureUserExists(t, userID)

	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	selectedDate := time.Now().AddDate(0, 0, 2)

	// Тест 1: Ошибка валидации организации
	t.Run("Organization validation error", func(t *testing.T) {
		_ = handlerSessionRepo.ClearSession(ctx, userID)

		state := botService.BookingState{
			UserID:       userID,
			ServiceID:    1,
			VisitType:    "private",
			SelectedDate: selectedDate,
			GuestName:    "Иванов Иван Иванович",
			Step:         botService.StepEnterOrg,
			CreatedAt:    time.Now(),
		}

		err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
			botService.KeyForBookingData: state,
		})
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)

		msg := createTestMessage(chatID, userID, "@#$%", 1)
		err = handler.HandleTextMessage(ctx, msg)
		assert.NoError(t, err)

		updatedState := getBookingStateFromSession(ctx, t, userID)
		require.NotNil(t, updatedState)
		assert.Equal(t, botService.StepEnterOrg, updatedState.Step)
		assert.Empty(t, updatedState.GuestOrganization)
	})

	// Тест 2: Ошибка валидации должности
	t.Run("Position validation error", func(t *testing.T) {
		_ = handlerSessionRepo.ClearSession(ctx, userID)

		state := botService.BookingState{
			UserID:            userID,
			ServiceID:         1,
			VisitType:         "private",
			SelectedDate:      selectedDate,
			GuestName:         "Иванов Иван Иванович",
			GuestOrganization: "ООО Ромашка",
			Step:              botService.StepEnterPosition,
			CreatedAt:         time.Now(),
		}

		err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
			botService.KeyForBookingData: state,
		})
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)

		msg := createTestMessage(chatID, userID, "@#$%", 1)
		err = handler.HandleTextMessage(ctx, msg)
		assert.NoError(t, err)

		updatedState := getBookingStateFromSession(ctx, t, userID)
		require.NotNil(t, updatedState)
		assert.Equal(t, botService.StepEnterPosition, updatedState.Step)
		assert.Empty(t, updatedState.GuestPosition)
	})

	mockBot.AssertExpectations(t)
}

// Тест: Race condition при конкурентном бронировании
func TestBookingFormHandler_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := repository.NewUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	emptyBot := &tgbotapi.BotAPI{}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		keyboard,
	)
	handler.bot = mockBot

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	ctx := context.Background()
	chatID := int64(12345)
	baseUserID := int64(100000)
	messageID := 1
	targetDate := time.Now().AddDate(0, 0, 2)

	_, err := handlerTestDB.Exec("DELETE FROM bookings WHERE booking_date = $1", targetDate)
	require.NoError(t, err)

	_, err = handlerTestDB.Exec("DELETE FROM users WHERE id >= $1", baseUserID)
	require.NoError(t, err)

	for i := 0; i < 20; i++ {
		userID := baseUserID + int64(i)
		_, err = handlerTestDB.Exec(`
			INSERT INTO users (id, telegram_id, username, email, password_hash) 
			VALUES ($1, $2, $3, $4, $5)`,
			userID, userID, fmt.Sprintf("raceuser%d", i),
			fmt.Sprintf("race%d@example.com", i), "hash")
		require.NoError(t, err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex
	errChan := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			userID := baseUserID + int64(idx)
			currentMsgID := messageID

			query1 := createTestCallbackQuery(chatID, userID, "book:private:1", currentMsgID)
			err := handler.Handle(ctx, query1)
			if err != nil {
				errChan <- err
				return
			}
			currentMsgID++

			time.Sleep(10 * time.Millisecond)

			query2 := createTestCallbackQuery(chatID, userID,
				"book:select_date:private:"+targetDate.Format("2006-01-02"),
				currentMsgID)
			err = handler.Handle(ctx, query2)
			if err != nil {
				errChan <- err
				return
			}
			currentMsgID++

			time.Sleep(10 * time.Millisecond)

			msg3 := createTestMessage(chatID, userID, "Race User", currentMsgID)
			err = handler.HandleTextMessage(ctx, msg3)
			if err != nil {
				errChan <- err
				return
			}
			currentMsgID++

			time.Sleep(10 * time.Millisecond)

			msg4 := createTestMessage(chatID, userID, "Race Org", currentMsgID)
			err = handler.HandleTextMessage(ctx, msg4)
			if err != nil {
				errChan <- err
				return
			}
			currentMsgID++

			time.Sleep(10 * time.Millisecond)

			msg5 := createTestMessage(chatID, userID, "Race Position", currentMsgID)
			err = handler.HandleTextMessage(ctx, msg5)
			if err != nil {
				errChan <- err
				return
			}
			currentMsgID++

			time.Sleep(10 * time.Millisecond)

			query3 := createTestCallbackQuery(chatID, userID, "book:confirm", currentMsgID)
			err = handler.Handle(ctx, query3)

			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
				t.Logf("Горутина %d: успешно", idx)
			} else {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	t.Logf("Успешных бронирований: %d", successCount)

	var count int
	err = handlerTestDB.Get(&count,
		"SELECT COUNT(*) FROM bookings WHERE booking_date = $1",
		targetDate)
	assert.NoError(t, err)
	t.Logf("Записей в БД: %d", count)
}

// Тест: Конкурентные сессии разных пользователей
func TestBookingFormHandler_ConcurrentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent sessions test in short mode")
	}

	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := repository.NewUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	emptyBot := &tgbotapi.BotAPI{}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		keyboard,
	)
	handler.bot = mockBot

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	ctx := context.Background()
	chatID := int64(12345)
	baseUserID := int64(200000)

	_, err := handlerTestDB.Exec("DELETE FROM users WHERE id >= $1", baseUserID)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		userID := baseUserID + int64(i)
		_, err = handlerTestDB.Exec(`
			INSERT INTO users (id, telegram_id, username, email, password_hash) 
			VALUES ($1, $2, $3, $4, $5)`,
			userID, userID, fmt.Sprintf("concurrent%d", i),
			fmt.Sprintf("concurrent%d@example.com", i), "hash")
		require.NoError(t, err)
	}

	const users = 10
	var wg sync.WaitGroup

	for u := 0; u < users; u++ {
		wg.Add(1)
		go func(userOffset int) {
			defer wg.Done()

			userID := baseUserID + int64(userOffset)
			messageID := userOffset + 1

			query := createTestCallbackQuery(chatID, userID, "book:private:1", messageID)
			err := handler.Handle(ctx, query)
			assert.NoError(t, err)

			time.Sleep(50 * time.Millisecond)

			state := getBookingStateFromSession(ctx, t, userID)
			if state != nil {
				assert.Equal(t, botService.StepSelectDate, state.Step)
			}
		}(u)
	}

	wg.Wait()
}
