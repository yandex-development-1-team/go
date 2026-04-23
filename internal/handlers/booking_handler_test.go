package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
	"github.com/yandex-development-1-team/go/internal/repository/postgres"
	reporedis "github.com/yandex-development-1-team/go/internal/repository/redis"
	"github.com/yandex-development-1-team/go/internal/service"
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
	bsService          *service.BoxSolutionsService
)

// TestUserID — константа для тестового пользователя
const TestUserID = int64(12345)

func TestMain(m *testing.M) {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	logger.NewLogger("dev", "debug")
	metrics.Initialize(config.Config{Environment: "test", HostName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	postgresContainer, err := startPostgresContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = initPostgresDB(postgresContainer)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %s", err.Error())
	}

	redisContainer, err := startRedisContainer(ctx)
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

	handlerService = botService.NewBookingService(handlerSessionRepo, handlerBookingRepo, handlerBoxRepo)
	bsService = service.NewBoxSolutionsService(handlerBoxRepo)

	code := m.Run()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = postgresContainer.Terminate(shutdownCtx)
	_ = redisContainer.Terminate(shutdownCtx)
	shutdownCancel()
	os.Exit(code)
}

func startPostgresContainer(ctx context.Context) (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL(nat.Port("5432/tcp"), "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(180 * time.Second),
	}

	return tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            false,
	})
}

func startRedisContainer(ctx context.Context) (tc.Container, error) {
	req := tc.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(120 * time.Second),
	}

	return tc.GenericContainer(ctx, tc.GenericContainerRequest{
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

	// Вставка тестовой услуги
	_, err = handlerTestDB.Exec(`
		INSERT INTO services (id, name, slug, description, rules, location, price, image, status, organizer) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING`,
		1, "Тестовая услуга 1", "test-service-1", "Описание 1", "Правила 1", "Москва", 1000, "", "active", "Тестовый организатор")
	if err != nil {
		return err
	}

	// Вставка тестовых слотов
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, 2),
		now.AddDate(0, 0, 3),
	}

	for _, date := range dates {
		_, err = handlerTestDB.Exec(`
			INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (service_id, slot_date, start_time, end_time) DO NOTHING`,
			1, date, "10:00:00", "12:00:00")
		if err != nil {
			return err
		}
		_, err = handlerTestDB.Exec(`
			INSERT INTO service_available_slots (service_id, slot_date, start_time, end_time) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (service_id, slot_date, start_time, end_time) DO NOTHING`,
			1, date, "14:00:00", "16:00:00")
		if err != nil {
			return err
		}
	}

	return nil
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

func createTestMessage(chatID int64, userID int64, text string, messageID int) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID: messageID,
		Chat:      &tgbotapi.Chat{ID: chatID},
		From:      &tgbotapi.User{ID: userID, UserName: "testuser"},
		Text:      text,
	}
}

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

// formatTimeForCallback преобразует время из формата HH:MM:SS в HH.MM.SS для callback
func formatTimeForCallback(timeStr string) string {
	return strings.ReplaceAll(timeStr, ":", ".")
}

// Тест 1: Начало бронирования
func TestBookingFormHandler_StartBooking(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 1

	ensureUserExists(t, userID)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil)

	query := createTestCallbackQuery(chatID, userID, "book:1:TestName:1", messageID)

	err := handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(300 * time.Millisecond)

	state := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, state)
	assert.Equal(t, userID, state.UserID)
	assert.Equal(t, int64(1), state.ServiceID)
	assert.Equal(t, botService.StepSelectDate, state.Step)

	mockBot.AssertExpectations(t)
}

// Тест 2: Выбор даты
func TestBookingFormHandler_SelectDate(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
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
		Step:      botService.StepSelectDate,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	// Преобразуем время для callback (заменяем : на .)
	startTimeCallback := formatTimeForCallback(startTime)
	endTimeCallback := formatTimeForCallback(endTime)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil)

	query := createTestCallbackQuery(chatID, userID, fmt.Sprintf("book:select_date:%s:%s:%s", targetDate, startTimeCallback, endTimeCallback), messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepEnterName, updatedState.Step)
	assert.Equal(t, targetDate, updatedState.SelectedSlot.Date)
	assert.Equal(t, startTime, updatedState.SelectedSlot.StartTime)
	assert.Equal(t, endTime, updatedState.SelectedSlot.EndTime)

	mockBot.AssertExpectations(t)
}

// Тест 3: Выбор недоступной даты
func TestBookingFormHandler_SelectUnavailableDate(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
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
		Step:      botService.StepSelectDate,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	unavailableDate := "2025-01-01"
	startTime := "10:00"
	endTime := "12:00"

	startTimeCallback := formatTimeForCallback(startTime)
	endTimeCallback := formatTimeForCallback(endTime)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{MessageID: messageID + 1, Chat: &tgbotapi.Chat{ID: chatID}}, nil).Once()

	query := createTestCallbackQuery(chatID, userID, fmt.Sprintf("book:select_date:%s:%s:%s", unavailableDate, startTimeCallback, endTimeCallback), messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)
	assert.Equal(t, botService.StepSelectDate, updatedState.Step)
	assert.Empty(t, updatedState.SelectedSlot.Date)

	mockBot.AssertExpectations(t)
}

// Тест 4: Ввод ФИО - успешный
func TestBookingFormHandler_EnterName_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 4

	ensureUserExists(t, userID)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	slot := models.BoxAvailableSlot{
		Date:      targetDate,
		StartTime: startTime,
		EndTime:   endTime,
	}
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		SelectedSlot: slot,
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
	assert.Equal(t, targetDate, updatedState.SelectedSlot.Date)

	mockBot.AssertExpectations(t)
}

// Тест 5: Ошибка валидации ФИО
func TestBookingFormHandler_EnterName_ValidationError(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 5

	ensureUserExists(t, userID)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	slot := models.BoxAvailableSlot{
		Date:      targetDate,
		StartTime: startTime,
		EndTime:   endTime,
	}
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		SelectedSlot: slot,
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
	assert.Equal(t, botService.StepEnterName, updatedState.Step)
	assert.Empty(t, updatedState.GuestName)
	assert.Equal(t, targetDate, updatedState.SelectedSlot.Date)

	mockBot.AssertExpectations(t)
}

// Тест 6: Ввод организации - успешный
func TestBookingFormHandler_EnterOrganization_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 6

	ensureUserExists(t, userID)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	slot := models.BoxAvailableSlot{
		Date:      targetDate,
		StartTime: startTime,
		EndTime:   endTime,
	}
	initialState := botService.BookingState{
		UserID:       userID,
		ServiceID:    1,
		SelectedSlot: slot,
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

	mockBot.AssertExpectations(t)
}

// Тест 7: Ввод должности - успешный
func TestBookingFormHandler_EnterPosition_Success(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 7

	ensureUserExists(t, userID)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	slot := models.BoxAvailableSlot{
		Date:      targetDate,
		StartTime: startTime,
		EndTime:   endTime,
	}
	initialState := botService.BookingState{
		UserID:            userID,
		ServiceID:         1,
		SelectedSlot:      slot,
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

	mockBot.AssertExpectations(t)
}

// Тест 8: Подтверждение бронирования
func TestBookingFormHandler_Confirmation(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 8

	ensureUserExists(t, userID)

	targetDate := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	startTime := "10:00"
	endTime := "12:00"

	slot := models.BoxAvailableSlot{
		Date:      targetDate,
		StartTime: startTime,
		EndTime:   endTime,
	}
	initialState := botService.BookingState{
		UserID:            userID,
		ServiceID:         1,
		ServiceName:       "TestName",
		SelectedSlot:      slot,
		GuestName:         "Иванов Иван Иванович",
		GuestOrganization: "ООО Ромашка",
		GuestPosition:     "Директор",
		Step:              botService.StepConfirmation,
		//OldMessageID:      0,
		CreatedAt: time.Now(),
	}

	err := handlerSessionRepo.SaveSession(ctx, userID, botService.CallbackBookingPrefix, map[string]interface{}{
		botService.KeyForBookingData: initialState,
	})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	mockBot.On("Request", mock.Anything).Return(&tgbotapi.APIResponse{Ok: true}, nil)
	mockBot.On("Send", mock.Anything).Return(tgbotapi.Message{}, nil).Times(1)

	query := createTestCallbackQuery(chatID, userID, "book:confirm", messageID)

	err = handler.Handle(ctx, query)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	updatedState := getBookingStateFromSession(ctx, t, userID)
	require.NotNil(t, updatedState)

	mockBot.AssertExpectations(t)
}

// Тест 9: Возврат в главное меню
func TestBookingFormHandler_BackToMainMenu(t *testing.T) {
	mockBot := new(MockBotAPI)
	emptyBot := &tgbotapi.BotAPI{}
	keyboard := NewKeyboardService()
	bsHandler := NewBoxSolutions(emptyBot, bsService)

	userRepo := postgres.NewTelegramUserRepository(handlerTestDB)
	startHandler := &StartHandler{
		bot:            mockBot,
		userRepository: userRepo,
	}

	handler := NewBookingFormHandler(
		emptyBot,
		handlerService,
		startHandler,
		bsHandler,
		keyboard,
	)
	handler.bot = mockBot

	ctx := context.Background()
	chatID := int64(12345)
	userID := TestUserID
	messageID := 9

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

	session, _ := handlerSessionRepo.GetSession(ctx, userID)
	assert.NotNil(t, session)

	mockBot.AssertExpectations(t)
}
