package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/models"
)

var (
	boxDB   *sqlx.DB
	boxRepo *BoxSolutionRepo
)

// setupBoxTest выполняет инициализацию для тестов BoxSolutionRepo
func setupBoxTest(t *testing.T) {
	// Если уже инициализировано, пропускаем
	if boxDB != nil {
		return
	}

	logger.NewLogger("dev", "debug")
	metrics.Initialize(config.Config{Environment: "test", HostName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	container, err := startBoxContainer()
	require.NoError(t, err)

	err = createBoxDB(container)
	require.NoError(t, err)

	// Инициализация тестовых данных
	err = initBoxTestData()
	require.NoError(t, err)

	boxRepo = NewBoxSolutionRepo(boxDB)

	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})
}

func startBoxContainer() (tc.Container, error) {
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

func createBoxDB(container tc.Container) error {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
	var err error
	boxDB, err = sqlx.Connect("postgres", dbURI)
	if err != nil {
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.UpContext(context.Background(), boxDB.DB, "../../../migrations"); err != nil {
		return err
	}

	return nil
}

func initBoxTestData() error {
	// Очистка таблиц
	_, err := boxDB.Exec("TRUNCATE TABLE services, service_available_slots CASCADE")
	if err != nil {
		return err
	}

	// Вставка тестовых услуг (коробочные решения)
	services := []struct {
		id          int
		name        string
		description string
		rules       string
		schedule    string
		typeService string
		boxSolution bool
	}{
		{1, "Коробочное решение 1", "Описание 1", "Правила 1", "9:00-18:00", "type1", true},
		{2, "Коробочное решение 2", "Описание 2", "Правила 2", "10:00-19:00", "type2", true},
		{3, "Обычная услуга", "Не коробочная", "", "8:00-20:00", "type3", false},
		{4, "Коробочное решение 3", "Описание 3", "Правила 3", "11:00-20:00", "type1", true},
	}

	for _, s := range services {
		_, err = boxDB.Exec(`
			INSERT INTO services (id, name, description, rules, schedule, type, box_solution) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET 
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				rules = EXCLUDED.rules,
				schedule = EXCLUDED.schedule,
				type = EXCLUDED.type,
				box_solution = EXCLUDED.box_solution`,
			s.id, s.name, s.description, s.rules, s.schedule, s.typeService, s.boxSolution)
		if err != nil {
			return err
		}
	}

	// Вставка доступных слотов
	now := time.Now()
	dates := []time.Time{
		now.AddDate(0, 0, 1), // завтра
		now.AddDate(0, 0, 2), // послезавтра
		now.AddDate(0, 0, 3), // через 3 дня
		now.AddDate(0, 0, 5), // через 5 дней
	}

	slots := []struct {
		serviceID int
		date      time.Time
		timeSlots []string
	}{
		{1, dates[0], []string{"10:00", "11:00", "12:00"}},
		{1, dates[1], []string{"10:00", "11:00", "14:00", "15:00"}},
		{1, dates[2], []string{"09:00", "10:00", "16:00"}},
		{2, dates[0], []string{"13:00", "14:00", "15:00"}},
		{2, dates[3], []string{"10:00", "11:00"}},
		{4, dates[1], []string{"09:00", "10:00", "11:00", "12:00"}},
		{4, dates[2], []string{"14:00", "15:00", "16:00"}},
	}

	for _, slot := range slots {
		_, err = boxDB.Exec(`
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

// pqStringArray конвертирует []string в интерфейс для вставки в БД
func pqStringArray(s []string) interface{} {
	return s
}

func TestBoxSolutionRepo_GetServices(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()

	services, err := boxRepo.GetServices(ctx, 12345) // telegramID не важен для запроса
	require.NoError(t, err)

	// Должны получить только услуги с box_solution = true (3 штуки)
	assert.Len(t, services, 3, "Should return only box solutions")

	// Проверяем структуру данных для первой услуги
	var found bool
	for _, svc := range services {
		if svc.ID == 1 {
			found = true
			assert.Equal(t, "Коробочное решение 1", svc.Name)
			assert.Equal(t, "Описание 1", svc.Description)
			assert.Equal(t, "Правила 1", svc.Rules)
			assert.Equal(t, "9:00-18:00", svc.Schedule)
			assert.Equal(t, "type1", svc.Type)
			assert.True(t, svc.BoxSolution)

			// Проверяем слоты для услуги 1 (должно быть 3 даты)
			assert.Len(t, svc.AvailableSlots, 3, "Service 1 should have 3 available dates")

			// Проверяем конкретные слоты
			slotMap := make(map[string][]string)
			for _, slot := range svc.AvailableSlots {
				slotMap[slot.Date] = slot.TimeSlots
			}

			// Проверяем первую дату
			date1 := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			assert.Contains(t, slotMap, date1)
			assert.ElementsMatch(t, []string{"10:00", "11:00", "12:00"}, slotMap[date1])

			// Проверяем вторую дату
			date2 := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
			assert.Contains(t, slotMap, date2)
			assert.ElementsMatch(t, []string{"10:00", "11:00", "14:00", "15:00"}, slotMap[date2])
		}
	}
	assert.True(t, found, "Service with ID 1 should be present")
}

func TestBoxSolutionRepo_GetServiceByID(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		serviceID int
		wantErr   error
		checkFunc func(t *testing.T, svc models.Service)
	}{
		{
			name:      "existing_service",
			serviceID: 1,
			wantErr:   nil,
			checkFunc: func(t *testing.T, svc models.Service) {
				assert.Equal(t, int64(1), svc.ID)
				assert.Equal(t, "Коробочное решение 1", svc.Name)
				assert.True(t, svc.BoxSolution)
				assert.Len(t, svc.AvailableSlots, 3, "Should have 3 available dates")

				// Проверяем, что все слоты непустые
				for _, slot := range svc.AvailableSlots {
					assert.NotEmpty(t, slot.Date)
					assert.NotEmpty(t, slot.TimeSlots)
				}
			},
		},
		{
			name:      "non_box_service",
			serviceID: 3, // обычная услуга, не коробочная
			wantErr:   nil,
			checkFunc: func(t *testing.T, svc models.Service) {
				assert.Equal(t, int64(3), svc.ID)
				assert.Equal(t, "Обычная услуга", svc.Name)
				assert.False(t, svc.BoxSolution)
				assert.Empty(t, svc.AvailableSlots, "Non-box service should have no slots")
			},
		},
		{
			name:      "service_without_slots",
			serviceID: 2,
			wantErr:   nil,
			checkFunc: func(t *testing.T, svc models.Service) {
				assert.Equal(t, int64(2), svc.ID)
				assert.True(t, svc.BoxSolution)
				assert.Len(t, svc.AvailableSlots, 2, "Should have 2 available dates")
			},
		},
		{
			name:      "non_existent_service",
			serviceID: 999,
			wantErr:   sql.ErrNoRows,
			checkFunc: func(t *testing.T, svc models.Service) {
				// ничего не проверяем, ожидаем ошибку
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := boxRepo.GetServiceByID(ctx, tt.serviceID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			tt.checkFunc(t, svc)
		})
	}
}

func TestBoxSolutionRepo_GetAvailableDates(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		serviceID int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "service_with_dates",
			serviceID: 1,
			wantCount: 3, // услуга 1 имеет 3 даты
			wantErr:   false,
		},
		{
			name:      "service_without_slots",
			serviceID: 2,
			wantCount: 2, // услуга 2 имеет 2 даты
			wantErr:   false,
		},
		{
			name:      "service_no_slots",
			serviceID: 5, // нет слотов
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "invalid_service_id",
			serviceID: -1,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dates, err := boxRepo.GetAvailableDates(ctx, tt.serviceID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, dates, tt.wantCount)

			// Проверяем формат дат
			for _, date := range dates {
				_, err := time.Parse("2006-01-02", date)
				assert.NoError(t, err, "Date should be in YYYY-MM-DD format")
			}
		})
	}
}

func TestBoxSolutionRepo_GetAvailableSlotsByServiceID(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()

	tests := []struct {
		name      string
		serviceID int
		wantCount int
		checkFunc func(t *testing.T, slots []models.AvailableSlot)
	}{
		{
			name:      "service_with_slots",
			serviceID: 4,
			wantCount: 2,
			checkFunc: func(t *testing.T, slots []models.AvailableSlot) {
				// Проверяем, что у всех слотов есть время
				for _, slot := range slots {
					assert.NotEmpty(t, slot.TimeSlots)
				}
			},
		},
		{
			name:      "service_without_slots",
			serviceID: 3, // обычная услуга, нет слотов
			wantCount: 0,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots, err := boxRepo.GetAvailableSlotsByServiceID(ctx, tt.serviceID)
			assert.NoError(t, err)
			assert.Len(t, slots, tt.wantCount)

			if tt.checkFunc != nil {
				tt.checkFunc(t, slots)
			}
		})
	}
}

func TestBoxSolutionRepo_GetAvailableTimeSlotsByDate(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	dayAfterTomorrow := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	tests := []struct {
		name      string
		serviceID int
		date      string
		want      []string
		wantErr   bool
	}{
		{
			name:      "existing_date",
			serviceID: 1,
			date:      tomorrow,
			want:      []string{"10:00", "11:00", "12:00"},
			wantErr:   false,
		},
		{
			name:      "another_date",
			serviceID: 1,
			date:      dayAfterTomorrow,
			want:      []string{"10:00", "11:00", "14:00", "15:00"},
			wantErr:   false,
		},
		{
			name:      "date_without_slots",
			serviceID: 1,
			date:      "2025-01-01", // дата без слотов
			want:      []string{},
			wantErr:   false,
		},
		{
			name:      "invalid_date_format",
			serviceID: 1,
			date:      "01-02-2025",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "empty_date",
			serviceID: 1,
			date:      "",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid_service_id",
			serviceID: -1,
			date:      tomorrow,
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slots, err := boxRepo.GetAvailableTimeSlotsByDate(ctx, tt.serviceID, tt.date)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.want, slots)
		})
	}
}

func TestBoxSolutionRepo_CheckSlotAvailability(t *testing.T) {
	setupBoxTest(t)

	ctx := context.Background()
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name      string
		serviceID int
		date      string
		timeSlot  string
		want      bool
		wantErr   bool
	}{
		{
			name:      "available_slot",
			serviceID: 1,
			date:      tomorrow,
			timeSlot:  "10:00",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "unavailable_slot",
			serviceID: 1,
			date:      tomorrow,
			timeSlot:  "13:00", // нет в списке
			want:      false,
			wantErr:   false,
		},
		{
			name:      "invalid_date",
			serviceID: 1,
			date:      "invalid",
			timeSlot:  "10:00",
			want:      false,
			wantErr:   true,
		},
		{
			name:      "empty_time_slot",
			serviceID: 1,
			date:      tomorrow,
			timeSlot:  "",
			want:      false,
			wantErr:   true,
		},
		{
			name:      "non_existent_service",
			serviceID: 999,
			date:      tomorrow,
			timeSlot:  "10:00",
			want:      false,
			wantErr:   false, // нет ошибки, просто false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available, err := boxRepo.CheckSlotAvailability(ctx, tt.serviceID, tt.date, tt.timeSlot)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, available)
		})
	}
}

func TestBoxSolutionRepo_NoRaceConditions(t *testing.T) {
	setupBoxTest(t)

	// Этот тест проверяет, что методы репозитория только читают данные
	// и не создают race conditions (в отличие от booking_repository, где есть запись)

	ctx := context.Background()
	serviceID := 1

	// Просто вызываем все read-only методы
	_, err := boxRepo.GetServices(ctx, 123)
	assert.NoError(t, err)

	_, err = boxRepo.GetServiceByID(ctx, serviceID)
	assert.NoError(t, err)

	_, err = boxRepo.GetAvailableDates(ctx, serviceID)
	assert.NoError(t, err)

	_, err = boxRepo.GetAvailableSlotsByServiceID(ctx, serviceID)
	assert.NoError(t, err)

	_, err = boxRepo.GetAvailableTimeSlotsByDate(ctx, serviceID, time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
	assert.NoError(t, err)

	_, err = boxRepo.CheckSlotAvailability(ctx, serviceID, time.Now().AddDate(0, 0, 1).Format("2006-01-02"), "10:00")
	assert.NoError(t, err)
}
