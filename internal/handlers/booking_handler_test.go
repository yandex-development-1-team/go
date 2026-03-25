package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/repository/postgres"
	reporedis "github.com/yandex-development-1-team/go/internal/repository/redis"
	botService "github.com/yandex-development-1-team/go/internal/service/bot"
)

// MockBotAPI — мок для Telegram Bot API
type MockBotAPI struct {
	mock.Mock
}

func (m *MockBotAPI) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return tgbotapi.Message{}, args.Error(1)
	}
	return args.Get(0).(tgbotapi.Message), args.Error(1)
}

func (m *MockBotAPI) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tgbotapi.APIResponse), args.Error(1)
}

var (
	handlerTestDB      *sqlx.DB
	handlerTestRedis   *redis.Client
	handlerSessionRepo repository.SessionRepository
	handlerBookingRepo *postgres.BookingRepo
	handlerBoxRepo     *postgres.BoxSolutionRepo
	handlerService     *botService.BookingService
)

// TestUserID — константа для тестового пользователя
const TestUserID = int64(12345)

func TestMain(m *testing.M) {
	logger.NewLogger("dev", "debug")
	metrics.Initialize(config.Config{Environment: "test", HostName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	postgresContainer, err := startPostgresContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = initPostgresDB(postgresContainer)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %s", err.Error())
	}

	redisContainer, err := startRedisContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = initRedisClient(redisContainer)
	if err != nil {
		log.Fatalf("failed to connect to redis: %s", err.Error())
	}

	err = initHandlerTestData()
	if err != nil {
		log.Fatalf("failed to init test data: %s", err.Error())
	}

	handlerSessionRepo = reporedis.NewSessionRepository(handlerTestRedis)
	handlerBookingRepo = postgres.NewBookingRepository(handlerTestDB)
	handlerBoxRepo = postgres.NewBoxSolutionRepo(handlerTestDB)

	// Создаем сервис один раз для всех тестов
	handlerService = botService.NewBookingService(handlerSessionRepo, handlerBookingRepo, handlerBoxRepo)

	code := m.Run()

	_ = postgresContainer.Terminate(ctx)
	_ = redisContainer.Terminate(ctx)
	os.Exit(code)
}

func startPostgresContainer() (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL(nat.Port("5432/tcp"), "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(120 * time.Second),
	}

	return tc.GenericContainer(context.Background(), tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func startRedisContainer() (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	return tc.GenericContainer(context.Background(), tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func initPostgresDB(container tc.Container) error {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
	var err error
	handlerTestDB, err = sqlx.Connect("postgres", dbURI)
	if err != nil {
		return err
	}

	migrationsPath, err := database.ResolveMigrationsDir("")
	if err != nil {
		return fmt.Errorf("migrations dir: %w", err)
	}
	if err := database.RunMigrations(handlerTestDB.DB, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	handlerTestDB.SetMaxOpenConns(50)
	handlerTestDB.SetMaxIdleConns(20)

	return nil
}

func initRedisClient(container tc.Container) error {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "6379")

	handlerTestRedis = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	return handlerTestRedis.Ping(context.Background()).Err()
}

func initHandlerTestData() error {
	_, err := handlerTestDB.Exec("TRUNCATE TABLE services, users, bookings, service_available_slots CASCADE")
	if err != nil {
		return err
	}

	_, err = handlerTestDB.Exec(`
		INSERT INTO users (id, telegram_id, username, email, password_hash) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING`,
		TestUserID, TestUserID, "testuser", "test@example.com", "hash")
	if err != nil {
		return err
	}

	services := []struct {
		id          int
		name        string
		description string
		rules       string
		schedule    string
		typeService string
		boxSolution bool
	}{
		{1, "Тестовая услуга 1", "Описание 1", "Правила 1", "9:00-18:00", "museum", true},
	}

	for _, s := range services {
		_, err = handlerTestDB.Exec(`
			INSERT INTO services (id, name, description, rules, schedule, type, box_solution) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING`,
			s.id, s.name, s.description, s.rules, s.schedule, s.typeService, s.boxSolution)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, 2),
		now.AddDate(0, 0, 3),
	}

	slots := []struct {
		serviceID int
		date      time.Time
		timeSlots []string
	}{
		{1, dates[0], []string{"10:00", "11:00", "12:00"}},
		{1, dates[1], []string{"10:00", "11:00", "14:00", "15:00"}},
	}

	for _, slot := range slots {
		_, err = handlerTestDB.Exec(`
			INSERT INTO service_available_slots (service_id, slot_date, time_slots) 
			VALUES ($1, $2, $3)
			ON CONFLICT (service_id, slot_date) DO UPDATE SET time_slots = EXCLUDED.time_slots`,
			slot.serviceID, slot.date, pqStringArray(slot.timeSlots))
		if err != nil {
			return err
		}
	}

	return nil
}

func pqStringArray(s []string) interface{} {
	return s
}

func createTestCallbackQuery(chatID int64, userID int64, data string, messageID int) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID:   "test_callback_id",
		From: &tgbotapi.User{ID: userID, UserName: "testuser"},
		Message: &tgbotapi.Message{
			Chat:      &tgbotapi.Chat{ID: chatID},
			MessageID: messageID,
			From:      &tgbotapi.User{ID: userID, UserName: "testuser"},
		},
		Data: data,
	}
}

// createTestMessage создает тестовое сообщение
func createTestMessage(chatID int64, userID int64, text string, messageID int) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID: messageID,
		Chat:      &tgbotapi.Chat{ID: chatID},
		From:      &tgbotapi.User{ID: userID, UserName: "testuser"},
		Text:      text,
	}
}

// Вспомогательная функция для получения состояния из сессии
func getBookingStateFromSession(ctx context.Context, t *testing.T, userID int64) *botService.BookingState {
	session, err := handlerSessionRepo.GetSession(ctx, userID)
	if err != nil {
		return nil
	}

	stateData, ok := session.StateData[botService.KeyForBookingData]
	if !ok {
		return nil
	}

	jsonData, err := json.Marshal(stateData)
	require.NoError(t, err)

	var state botService.BookingState
	err = json.Unmarshal(jsonData, &state)
	require.NoError(t, err)

	return &state
}

// ensureUserExists гарантирует, что пользователь существует в БД
func ensureUserExists(t *testing.T, userID int64) {
	var exists bool
	err := handlerTestDB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
	require.NoError(t, err)

	if !exists {
		_, err = handlerTestDB.Exec(`
			INSERT INTO users (id, telegram_id, username, email, password_hash) 
			VALUES ($1, $2, $3, $4, $5)`,
			userID, userID, "testuser", "test@example.com", "hash")
		require.NoError(t, err)
	}
}

// Тест 1: Начало бронирования
func TestBookingFormHandler_StartBooking(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	// Используем реальный репозиторий вместо мока
	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)

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

	// Гарантируем, что пользователь существует
	ensureUserExists(t, userID)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil)

	query := createTestCallbackQuery(chatID, userID, "book:private:1", messageID)

	err := handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)

	state := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state)
	assert.Equal(t, userID, state.UserID)
	assert.Equal(t, 1, state.ServiceID)
	assert.Equal(t, "private", state.VisitType)
	assert.Equal(t, botService.StepSelectDate, state.Step)
	// LastMsgID не проверяем, так как он не сохраняется в этой версии

	mockBot.AssertExpectations(t)
}

// Тест 2: Выбор даты
func TestBookingFormHandler_SelectDate(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 2

	ensureUserExists(t, userID)

	initialState := botService.BookingState{
		UserID:    userID,
		ServiceID: 1,
		VisitType: "private",
		Step:      botService.StepSelectDate,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	targetDate := time.Now().AddDate(0, 0, 2)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil)

	query := createTestCallbackQuery(chatID, userID, "book:select_date:private:"+targetDate.Format("2006-01-02"), messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepEnterName, updatedState.Step)
	assert.Equal(t, targetDate.Format("2006-01-02"), updatedState.SelectedDate.Format("2006-01-02"))

	mockBot.AssertExpectations(t)
}

// Тест 3: Выбор недоступной даты
func TestBookingFormHandler_SelectUnavailableDate(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 3

	ensureUserExists(t, userID)

	initialState := botService.BookingState{
		UserID:    userID,
		ServiceID: 1,
		VisitType: "private",
		Step:      botService.StepSelectDate,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	unavailableDate := "2025-01-01"

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	// ТОЛЬКО ОДИН вызов Send - от startBooking при перезапуске
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil).Once()

	query := createTestCallbackQuery(chatID, userID, "book:select_date:private:"+unavailableDate, messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepSelectDate, updatedState.Step)
	assert.True(t, updatedState.SelectedDate.IsZero())

	mockBot.AssertExpectations(t)
}

// Тест 4: Ввод ФИО - успешный
func TestBookingFormHandler_EnterName_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 4

	ensureUserExists(t, userID)

	selectedDate := time.Now().AddDate(0, 0, 2)
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		VisitType:    "private",
		SelectedDate: selectedDate,
		Step:         botService.StepEnterName,
		CreatedAt:    time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	msg := createTestMessage(chatID, userID, "Иванов Иван Иванович", messageID)
	err = handler.HandleTextMessage(ctx, msg)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepEnterOrg, updatedState.Step)
	assert.Equal(t, "Иванов Иван Иванович", updatedState.GuestName)
	assert.Equal(t, selectedDate.Format("2006-01-02"), updatedState.SelectedDate.Format("2006-01-02"))

	mockBot.AssertExpectations(t)
}

// Тест 5: Ошибка валидации ФИО - состояние НЕ сохраняется
func TestBookingFormHandler_EnterName_ValidationError(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 5

	ensureUserExists(t, userID)

	selectedDate := time.Now().AddDate(0, 0, 2)
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		VisitType:    "private",
		SelectedDate: selectedDate,
		Step:         botService.StepEnterName,
		CreatedAt:    time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	msg := createTestMessage(chatID, userID, "123", messageID)
	err = handler.HandleTextMessage(ctx, msg)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	// Шаг не должен измениться
	assert.Equal(t, botService.StepEnterName, updatedState.Step)
	// Имя не должно сохраниться
	assert.Empty(t, updatedState.GuestName)
	assert.Equal(t, selectedDate.Format("2006-01-02"), updatedState.SelectedDate.Format("2006-01-02"))

	mockBot.AssertExpectations(t)
}

// Тест 6: Ввод организации - успешный
func TestBookingFormHandler_EnterOrganization_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 6

	ensureUserExists(t, userID)

	selectedDate := time.Now().AddDate(0, 0, 2)
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		VisitType:    "private",
		SelectedDate: selectedDate,
		GuestName:    "Иванов Иван Иванович",
		Step:         botService.StepEnterOrg,
		CreatedAt:    time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	msg := createTestMessage(chatID, userID, "ООО Ромашка", messageID)
	err = handler.HandleTextMessage(ctx, msg)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepEnterPosition, updatedState.Step)
	assert.Equal(t, "ООО Ромашка", updatedState.GuestOrganization)
	assert.Equal(t, "Иванов Иван Иванович", updatedState.GuestName)
	assert.Equal(t, selectedDate.Format("2006-01-02"), updatedState.SelectedDate.Format("2006-01-02"))

	mockBot.AssertExpectations(t)
}

// Тест 7: Ввод должности - успешный
func TestBookingFormHandler_EnterPosition_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 7

	ensureUserExists(t, userID)

	selectedDate := time.Now().AddDate(0, 0, 2)
	initialState := botService.BookingState{
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
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	msg := createTestMessage(chatID, userID, "Директор", messageID)
	err = handler.HandleTextMessage(ctx, msg)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepConfirmation, updatedState.Step)
	assert.Equal(t, "Директор", updatedState.GuestPosition)
	assert.Equal(t, "ООО Ромашка", updatedState.GuestOrganization)
	assert.Equal(t, "Иванов Иван Иванович", updatedState.GuestName)
	assert.Equal(t, selectedDate.Format("2006-01-02"), updatedState.SelectedDate.Format("2006-01-02"))

	mockBot.AssertExpectations(t)
}

// Тест 8: Подтверждение бронирования
func TestBookingFormHandler_Confirmation(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 8

	ensureUserExists(t, userID)

	selectedDate := time.Now().AddDate(0, 0, 2)
	initialState := botService.BookingState{
		UserID:            userID,
		ServiceID:         1,
		VisitType:         "private",
		SelectedDate:      selectedDate,
		GuestName:         "Иванов Иван Иванович",
		GuestOrganization: "ООО Ромашка",
		GuestPosition:     "Директор",
		Step:              botService.StepConfirmation,
		CreatedAt:         time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil).Times(2)

	query := createTestCallbackQuery(chatID, userID, "book:confirm", messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	session, _ := handlerSessionRepo.GetSession(ctx, userID)
	assert.Nil(t, session)

	mockBot.AssertExpectations(t)
}

// Тест 9: Возврат в главное меню
func TestBookingFormHandler_BackToMainMenu(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	// Используем реальный репозиторий вместо мока
	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)

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
	messageID := 9

	// Создаем пользователя в БД
	ensureUserExists(t, userID)

	initialState := botService.BookingState{
		UserID:    userID,
		ServiceID: 1,
		Step:      botService.StepEnterName,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	query := createTestCallbackQuery(chatID, userID, "book:main_menu", messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	// Проверяем, что сессия очищена
	session, _ := handlerSessionRepo.GetSession(ctx, userID)
	assert.Nil(t, session)

	mockBot.AssertExpectations(t)
}

// Тест 10: Неизвестное действие
func TestBookingFormHandler_UnknownAction(t *testing.T) {
	mockBot := new(MockBotAPI)
	keyboard := NewKeyboardService()

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
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
	messageID := 10

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil)

	query := createTestCallbackQuery(chatID, userID, "book:unknown", messageID)

	err := handler.Handle(ctx, query)
	assert.NoError(t, err)

	mockBot.AssertExpectations(t)
}
