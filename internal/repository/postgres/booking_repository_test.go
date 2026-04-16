package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/database"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/metrics"
	"github.com/yandex-development-1-team/go/internal/models"
)

var (
	db               *sqlx.DB
	repo             *BookingRepo
	repoUser         *TelegramUserRepo
	repoSession      *pgSessionRepo
	boxRepo          *BoxSolutionRepo
	resourcePageRepo *ResourcePageRepository
	applicationRepo  *ApplicationRepo
)

// cleanBookingsTables удаляет данные из всех связанных таблиц
func cleanBookingsTables(t *testing.T) {
	_, err := db.Exec("TRUNCATE TABLE bookings, users, services, staff RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

// seedUser создаёт пользователя с заданным telegram_id и username
func seedUser(t *testing.T, telegramID int64, username string) {
	_, err := db.Exec(`
			INSERT INTO users (telegram_id, username) VALUES ($1, $2)
			ON CONFLICT (telegram_id) DO NOTHING
	`, telegramID, username)
	require.NoError(t, err)
}

// seedService создаёт сервис
func seedService(t *testing.T, id int64, name string) {
	_, err := db.Exec(`
			INSERT INTO services (id, name, slug, description, price, status, organizer) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING
	`, id, name, "test-slug", "test description", 1000, "active", "Test Organizer")
	require.NoError(t, err)
}

// seedStaff создаёт сотрудника
func seedStaff(t *testing.T, id int64, email, role string) {
	_, err := db.Exec(`
			INSERT INTO staff (id, email, role, first_name, last_name) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
	`, id, email, role, "Test", "Manager")
	require.NoError(t, err)
}

// seedBooking вставляет бронирование и возвращает его ID
func seedBooking(t *testing.T, b *models.BookingAPI) int64 {
	bookingDate, err := time.Parse("2006-01-02", b.BookingDate)
	require.NoError(t, err)

	var bookingTime *time.Time
	if b.BookingTime != "" {
		tm, err := time.Parse("15:04", b.BookingTime)
		require.NoError(t, err)
		bookingTime = &tm
	}

	// Преобразуем ManagerID: 0 → nil (NULL), иначе указатель на значение
	var managerID interface{}
	if b.ManagerID != 0 {
		managerID = b.ManagerID
	} else {
		managerID = nil
	}

	var id int64
	err = db.QueryRow(`
			INSERT INTO bookings (
					user_id, service_id, booking_date, booking_time,
					guest_name, guest_organization, guest_position,
					status, manager_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
	`,
		b.UserID, b.ServiceID, bookingDate, bookingTime,
		b.GuestName, b.GuestOrganization, b.GuestPosition,
		b.Status, managerID, b.CreatedAt, b.UpdatedAt,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func mustParseTime(layout, value string) *time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time %s: %v", value, err))
	}
	return &t
}

func TestMain(m *testing.M) {
	logger.NewLogger("dev", "debug")
	metrics.Initialize(config.Config{Environment: "test", HostName: "test"})

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	container, err := startContainer()
	if err != nil {
		log.Fatal(err)
	}

	err = createDB(container)
	if err != nil {
		log.Fatalf("failed to connect to db: %s", err.Error())
	}

	_, _ = db.Exec(`INSERT INTO services (id, name, slug, price) VALUES (1, 'Test Service', 'test-service', 0) ON CONFLICT (id) DO NOTHING`)
	_, _ = db.Exec(`INSERT INTO services (id, name, slug, price) VALUES (5, 'Slot Service', 'slot-service', 0) ON CONFLICT (id) DO NOTHING`)
	_, _ = db.Exec(`INSERT INTO services (id, name, slug, price) VALUES (50, 'Race Service', 'race-service', 0) ON CONFLICT (id) DO NOTHING`)

	repo = NewBookingRepository(db)
	repoUser = NewTelegramUserRepository(db)
	repoSession = NewSessionRepository(db)
	boxRepo = NewBoxSolutionRepo(db)
	resourcePageRepo = NewResourcePageRepo(db)
	applicationRepo = NewApplicationRepository(db)

	code := m.Run()

	_ = container.Terminate(ctx)
	os.Exit(code)
}

func startContainer() (tc.Container, error) {
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

func createDB(container tc.Container) error {
	host, _ := container.Host(context.Background())
	port, _ := container.MappedPort(context.Background(), "5432")

	dbURI := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
	var err error
	db, err = sqlx.Connect("postgres", dbURI)
	if err != nil {
		return err
	}

	migDir, err := database.ResolveMigrationsDir("")
	if err != nil {
		return err
	}
	if err := database.RunMigrations(db.DB, migDir); err != nil {
		return err
	}

	return nil
}

func TestCreateBooking(t *testing.T) {
	// bookings.user_id FK → users.telegram_id (not users.id)
	const telegramID int64 = 111
	_, err := db.Exec(`
        INSERT INTO users (telegram_id, username, email, password_hash) VALUES ($1, $2, $3, $4)
        ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username`,
		telegramID, "test", "booking_test@test.local", "placeholder")
	assert.NoError(t, err)

	targetDate := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	targetTime := mustParseTime("15:04:05", "15:00:00")

	tests := []struct {
		name            string
		contextDuration time.Duration
		booking         *models.Booking
		wantErr         error
		preAction       func()
	}{
		{
			name:            "correct_data",
			contextDuration: 5 * time.Second,
			booking: &models.Booking{
				UserID:      telegramID,
				ServiceID:   1,
				BookingDate: targetDate,
				BookingTime: targetTime,
				GuestName:   "Tester",
			},
			preAction: func() {
				_, _ = db.Exec("DELETE FROM bookings")
			},
			wantErr: nil,
		},
		{
			name:            "request_canceled",
			contextDuration: 5 * time.Second,
			booking: &models.Booking{
				UserID:      telegramID,
				ServiceID:   1,
				BookingDate: targetDate,
				BookingTime: targetTime,
				GuestName:   "CanceledUser",
			},
			wantErr: models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Microsecond,
			booking: &models.Booking{
				UserID:      telegramID,
				ServiceID:   1,
				GuestName:   "TimeoutUser",
				BookingDate: targetDate,
				BookingTime: targetTime,
			},
			wantErr: models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preAction != nil {
				tt.preAction()
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			if tt.name == "request_canceled" {
				cancel()
			} else {
				defer cancel()
			}

			id, err := repo.CreateBooking(ctx, tt.booking)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Positive(t, id)
			}
		})
	}
}

func TestGetAvailableSlots(t *testing.T) {
	_, _ = db.Exec("DELETE FROM bookings")

	const telegramID int64 = 777
	_, err := db.Exec(`
        INSERT INTO users (telegram_id, username, email, password_hash) VALUES ($1, $2, $3, $4)
        ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username`,
		telegramID, "slot_tester", "slot_tester@test.local", "placeholder")
	assert.NoError(t, err)

	serviceID := 5
	targetDate := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)

	slotPending := "10:00:00"
	slotConfirmed := "11:00:00"

	_, err = db.Exec(`
        INSERT INTO bookings (user_id, service_id, booking_date, booking_time, guest_name, status)
        VALUES ($1, $2, $3, $4, 'Guest Pending', 'pending'),
               ($1, $2, $3, $5, 'Guest Confirmed', 'confirmed')`,
		telegramID, serviceID, targetDate, slotPending, slotConfirmed)
	assert.NoError(t, err)

	slots, err := repo.GetAvailableSlots(context.Background(), serviceID, targetDate)

	assert.NoError(t, err)
	if assert.Len(t, slots, 1, "Should only return slots where status is not confirmed") {
		assert.Equal(t, "10:00", slots[0].Format("15:04"))
	}
}

func TestGetBookingByID_Found(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1001)
	seedUser(t, telegramID, "testuser")
	seedService(t, 10, "Test Service")
	seedStaff(t, 20, "manager@test.com", "manager_1")

	// Проверяем, что менеджер действительно вставлен
	var email string
	err := db.Get(&email, "SELECT email FROM staff WHERE id = $1", 20)
	require.NoError(t, err)
	require.Equal(t, "manager@test.com", email)

	now := time.Now()
	booking := &models.BookingAPI{
		UserID:            telegramID,
		ServiceID:         10,
		BookingDate:       now.Format("2006-01-02"),
		BookingTime:       "14:30",
		GuestName:         "John Doe",
		GuestOrganization: "Test Org",
		GuestPosition:     "Engineer",
		Status:            "pending",
		ManagerID:         20,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	bookingID := seedBooking(t, booking)

	ctx := context.Background()
	got, err := repo.GetBookingById(ctx, bookingID)
	require.NoError(t, err)
	assert.Equal(t, bookingID, got.ID)
	assert.Equal(t, "Test Manager", got.ManagerName)
}

func TestGetBookingByID_NotFound(t *testing.T) {
	cleanBookingsTables(t)
	ctx := context.Background()
	_, err := repo.GetBookingById(ctx, 999999)
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}

func TestGetBookingByID_SoftDeleted(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1002)
	seedUser(t, telegramID, "softuser")
	seedService(t, 11, "Soft Service")

	now := time.Now()
	booking := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   11,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "10:00",
		GuestName:   "ToBeDeleted",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, booking)

	// Мягкое удаление
	_, err := db.Exec("UPDATE bookings SET deleted_at = NOW() WHERE id = $1", bookingID)
	require.NoError(t, err)

	ctx := context.Background()
	_, err = repo.GetBookingById(ctx, bookingID)
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}

func TestGetBookingsList_Empty(t *testing.T) {
	cleanBookingsTables(t)
	ctx := context.Background()
	filter := &models.ApplicationFilter{Limit: 20, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list.Items))
	assert.Equal(t, 0, list.Total)
}

func TestGetBookingsList_WithData(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1003)
	seedUser(t, telegramID, "listuser")
	seedService(t, 12, "List Service")
	seedStaff(t, 30, "listmgr@test.com", "manager_1")

	now := time.Now()
	for i := 0; i < 3; i++ {
		b := &models.BookingAPI{
			UserID:      telegramID,
			ServiceID:   12,
			BookingDate: now.Add(time.Duration(i) * 24 * time.Hour).Format("2006-01-02"),
			BookingTime: "12:00",
			GuestName:   "User",
			Status:      "pending",
			ManagerID:   30,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		seedBooking(t, b)
	}

	ctx := context.Background()
	filter := &models.ApplicationFilter{Limit: 10, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 3, len(list.Items))
	assert.Equal(t, 3, list.Total)
}

func TestGetBookingsList_FilterByStatus(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1004)
	seedUser(t, telegramID, "statususer")
	seedService(t, 13, "Status Service")
	seedStaff(t, 40, "statusmgr@test.com", "manager_1")

	now := time.Now()
	// pending
	b1 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   13,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "09:00",
		GuestName:   "Pending",
		Status:      "pending",
		ManagerID:   40,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b1)
	// confirmed
	b2 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   13,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "10:00",
		GuestName:   "Confirmed",
		Status:      "confirmed",
		ManagerID:   40,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b2)

	ctx := context.Background()
	filter := &models.ApplicationFilter{Status: "pending", Limit: 10, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list.Items))
	assert.Equal(t, 1, list.Total)
	assert.Equal(t, "pending", list.Items[0].Status)
}

func TestGetBookingsList_FilterByManagerID(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1005)
	seedUser(t, telegramID, "mgrfilter")
	seedService(t, 14, "Mgr Service")
	seedStaff(t, 50, "mgrA@test.com", "manager_1")
	seedStaff(t, 51, "mgrB@test.com", "manager_1")

	now := time.Now()
	b1 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   14,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "11:00",
		GuestName:   "MgrA",
		Status:      "pending",
		ManagerID:   50,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b1)
	b2 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   14,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "12:00",
		GuestName:   "MgrB",
		Status:      "pending",
		ManagerID:   51,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b2)

	ctx := context.Background()
	filter := &models.ApplicationFilter{ManagerID: 50, Limit: 10, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list.Items))
	assert.Equal(t, 1, list.Total)
	assert.Equal(t, int64(50), list.Items[0].ManagerID)
}

func TestGetBookingsList_FilterByCustomerName(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1006)
	seedUser(t, telegramID, "custfilter")
	seedService(t, 15, "Customer Service")
	seedStaff(t, 60, "custmgr@test.com", "manager_1")

	now := time.Now()
	b1 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   15,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "13:00",
		GuestName:   "Alice Johnson",
		Status:      "pending",
		ManagerID:   60,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b1)
	b2 := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   15,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "14:00",
		GuestName:   "Bob Smith",
		Status:      "pending",
		ManagerID:   60,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedBooking(t, b2)

	ctx := context.Background()
	filter := &models.ApplicationFilter{CustomerName: "alice", Limit: 10, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(list.Items))
	assert.Equal(t, 1, list.Total)
	assert.Contains(t, list.Items[0].GuestName, "Alice")
}

func TestGetBookingsList_Pagination(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1007)
	seedUser(t, telegramID, "paguser")
	seedService(t, 16, "Pagination Service")
	seedStaff(t, 70, "pagmgr@test.com", "manager_1")

	now := time.Now()
	for i := 0; i < 5; i++ {
		b := &models.BookingAPI{
			UserID:      telegramID,
			ServiceID:   16,
			BookingDate: now.Add(time.Duration(i) * 24 * time.Hour).Format("2006-01-02"),
			BookingTime: "15:00",
			GuestName:   "User",
			Status:      "pending",
			ManagerID:   70,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		seedBooking(t, b)
	}

	ctx := context.Background()
	filter1 := &models.ApplicationFilter{Limit: 2, Offset: 0}
	list1, err := repo.GetBookingsList(ctx, filter1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list1.Items))
	assert.Equal(t, 5, list1.Total)

	filter2 := &models.ApplicationFilter{Limit: 2, Offset: 2}
	list2, err := repo.GetBookingsList(ctx, filter2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(list2.Items))
	assert.NotEqual(t, list1.Items[0].ID, list2.Items[0].ID)
}

func TestDeleteBooking_Success(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1008)
	seedUser(t, telegramID, "deleteuser")
	seedService(t, 17, "Delete Service")

	now := time.Now()
	b := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   17,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "16:00",
		GuestName:   "ToDelete",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, b)

	ctx := context.Background()
	err := repo.DeleteBooking(ctx, bookingID)
	require.NoError(t, err)

	var deletedAt *time.Time
	err = db.QueryRow("SELECT deleted_at FROM bookings WHERE id = $1", bookingID).Scan(&deletedAt)
	require.NoError(t, err)
	require.NotNil(t, deletedAt)

	_, err = repo.GetBookingById(ctx, bookingID)
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}

func TestDeleteBooking_NotFound(t *testing.T) {
	cleanBookingsTables(t)
	ctx := context.Background()
	err := repo.DeleteBooking(ctx, 999999)
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}

func TestDeleteBooking_AlreadySoftDeleted(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(1009)
	seedUser(t, telegramID, "alreadydeleted")
	seedService(t, 18, "Already Deleted Service")

	now := time.Now()
	b := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   18,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "17:00",
		GuestName:   "SoftDeleted",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, b)

	_, err := db.Exec("UPDATE bookings SET deleted_at = NOW() WHERE id = $1", bookingID)
	require.NoError(t, err)

	var firstDeletedAt time.Time
	err = db.QueryRow("SELECT deleted_at FROM bookings WHERE id = $1", bookingID).Scan(&firstDeletedAt)
	require.NoError(t, err)

	// небольшая пауза чтобы NOW() дал другое значение
	time.Sleep(10 * time.Millisecond)

	ctx := context.Background()
	err = repo.DeleteBooking(ctx, bookingID)
	require.NoError(t, err) // WHERE id=$1 без фильтра deleted_at — всегда найдёт

	var secondDeletedAt time.Time
	err = db.QueryRow("SELECT deleted_at FROM bookings WHERE id = $1", bookingID).Scan(&secondDeletedAt)
	require.NoError(t, err)
	assert.True(t, secondDeletedAt.After(firstDeletedAt), "repeated soft delete must update deleted_at")
}

func TestGetBookingsList_SoftDeleted_NotIncluded(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(2001)
	seedUser(t, telegramID, "softlist")
	seedService(t, 20, "Soft List Service")

	now := time.Now()
	b := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   20,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "10:00",
		GuestName:   "SoftDeleted",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, b)

	_, err := db.Exec("UPDATE bookings SET deleted_at = NOW() WHERE id = $1", bookingID)
	require.NoError(t, err)

	ctx := context.Background()
	filter := &models.ApplicationFilter{Limit: 10, Offset: 0}
	list, err := repo.GetBookingsList(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 0, len(list.Items))
	assert.Equal(t, 0, list.Total)
}

func TestCreateBooking_InvalidUserID_ReturnsError(t *testing.T) {
	cleanBookingsTables(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b := &models.Booking{
		UserID:      999999999,
		ServiceID:   1,
		BookingDate: time.Now().AddDate(0, 0, 1),
		GuestName:   "Ghost",
	}
	_, err := repo.CreateBooking(ctx, b)
	require.Error(t, err)
	assert.NotErrorIs(t, err, models.ErrBookingNotFound)
}

func TestCreateBooking_InvalidServiceID_ReturnsError(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(2002)
	seedUser(t, telegramID, "fkservice")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b := &models.Booking{
		UserID:      telegramID,
		ServiceID:   9999,
		BookingDate: time.Now().AddDate(0, 0, 1),
		GuestName:   "Ghost",
	}
	_, err := repo.CreateBooking(ctx, b)
	require.Error(t, err)
	assert.NotErrorIs(t, err, models.ErrBookingNotFound)
}

func TestUpdateBookingStatus_Success(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(3001)
	seedUser(t, telegramID, "updatestatus")
	seedService(t, 30, "Update Status Service")

	now := time.Now()
	b := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   30,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "10:00",
		GuestName:   "ToUpdate",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, b)

	ctx := context.Background()
	err := repo.UpdateBookingStatus(ctx, bookingID, "confirmed")
	require.NoError(t, err)

	got, err := repo.GetBookingById(ctx, bookingID)
	require.NoError(t, err)
	assert.Equal(t, "confirmed", got.Status)
}

func TestUpdateBookingStatus_NotFound(t *testing.T) {
	cleanBookingsTables(t)

	ctx := context.Background()
	err := repo.UpdateBookingStatus(ctx, 999999, "confirmed")
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}

func TestUpdateBookingStatus_SoftDeleted_ReturnsNotFound(t *testing.T) {
	cleanBookingsTables(t)

	telegramID := int64(3002)
	seedUser(t, telegramID, "updatedeleted")
	seedService(t, 31, "Update Deleted Service")

	now := time.Now()
	b := &models.BookingAPI{
		UserID:      telegramID,
		ServiceID:   31,
		BookingDate: now.Format("2006-01-02"),
		BookingTime: "11:00",
		GuestName:   "SoftDeleted",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	bookingID := seedBooking(t, b)

	_, err := db.Exec("UPDATE bookings SET deleted_at = NOW() WHERE id = $1", bookingID)
	require.NoError(t, err)

	ctx := context.Background()
	err = repo.UpdateBookingStatus(ctx, bookingID, "confirmed")
	assert.ErrorIs(t, err, models.ErrBookingNotFound)
}
