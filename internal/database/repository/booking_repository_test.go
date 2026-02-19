package repository

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-development-1-team/go/internal/models"
)

// var (
// 	db   *sqlx.DB
// 	repo BookingRepository
// )

func mustParseTime(layout, value string) *time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time %s: %v", value, err))
	}
	return &t
}

// func TestMain(m *testing.M) {
// 	logger.NewLogger("dev", "debug")

// 	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
// 	defer cancel()

// 	container, err := startContainer()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = createDB(container)
// 	if err != nil {
// 		log.Fatalf("failed to connect to db: %s", err.Error())
// 	}

// 	repo = NewBookingRepository(db)

// 	code := m.Run()

// 	_ = container.Terminate(ctx)
// 	os.Exit(code)
// }

// func startContainer() (tc.Container, error) {
// 	req := tc.ContainerRequest{
// 		Image:        "postgres:latest",
// 		ExposedPorts: []string{"5432/tcp"},
// 		Env: map[string]string{
// 			"POSTGRES_PASSWORD": "password",
// 			"POSTGRES_DB":       "testdb",
// 		},
// 		WaitingFor: wait.ForSQL(nat.Port("5432/tcp"), "postgres", func(host string, port nat.Port) string {
// 			return fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
// 		}).WithStartupTimeout(120 * time.Second),
// 	}

// 	return tc.GenericContainer(context.Background(), tc.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// }

// func createDB(container tc.Container) error {
// 	host, _ := container.Host(context.Background())
// 	port, _ := container.MappedPort(context.Background(), "5432")

// 	dbURI := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
// 	var err error
// 	db, err = sqlx.Connect("postgres", dbURI)
// 	if err != nil {
// 		return err
// 	}

// 	if err := goose.SetDialect("postgres"); err != nil {
// 		return err
// 	}

// 	if err := goose.UpContext(context.Background(), db.DB, "../../../migrations"); err != nil {
// 		return err
// 	}

// 	return nil
// }

func TestCreateBooking(t *testing.T) {
	var userID int64
	err := db.QueryRow(`
        INSERT INTO users (telegram_id, username) VALUES ($1, $2)
        ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username
        RETURNING id`, 111, "test").Scan(&userID)
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
				UserID:      userID,
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
			name:            "duplicate_slot_error",
			contextDuration: 5 * time.Second,
			booking: &models.Booking{
				UserID:      userID,
				ServiceID:   1,
				BookingDate: targetDate,
				BookingTime: targetTime,
				GuestName:   "Other",
			},
			preAction: func() {
				_, _ = db.Exec(`INSERT INTO bookings (user_id, service_id, booking_date, booking_time, guest_name, status)
                    VALUES ($1, 1, $2, $3, 'Owner', 'confirmed')`, userID, targetDate, targetTime)
			},
			wantErr: models.ErrSlotOccupied,
		},
		{
			name:            "request_canceled",
			contextDuration: 5 * time.Second,
			booking: &models.Booking{
				UserID:      userID,
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
				UserID:      userID,
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

			id, err := repoBooking.CreateBooking(ctx, tt.booking)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Positive(t, id)
			}
		})
	}
}

func TestCreateBooking_RaceCondition(t *testing.T) {
	const goroutines = 10
	serviceID := 50
	date := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	slot := mustParseTime("15:04:05", "12:00:00")

	var userID int64
	_ = db.QueryRow("INSERT INTO users (telegram_id, username) VALUES (999, 'racer') ON CONFLICT DO NOTHING RETURNING id").Scan(&userID)
	if userID == 0 {
		_ = db.Get(&userID, "SELECT id FROM users WHERE telegram_id = 999")
	}

	_, _ = db.Exec("DELETE FROM bookings WHERE service_id = $1 AND booking_date = $2", serviceID, date)

	results := make(chan error, goroutines)
	var successCount int32

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := repoBooking.CreateBooking(context.Background(), &models.Booking{
				UserID:      userID,
				ServiceID:   int16(serviceID),
				BookingDate: date,
				BookingTime: slot,
				GuestName:   "Racer",
			})

			if err == nil {
				atomic.AddInt32(&successCount, 1)
				_, _ = db.Exec("UPDATE bookings SET status = 'confirmed' WHERE service_id = $1", serviceID)
			}
			results <- err
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-results
	}

	assert.Equal(t, int32(1), successCount, "Only one booking should be successful")
}

func TestGetAvailableSlots(t *testing.T) {
	_, _ = db.Exec("DELETE FROM bookings")

	var userID int64
	err := db.QueryRow(`
        INSERT INTO users (telegram_id, username) VALUES (777, 'slot_tester')
        ON CONFLICT (telegram_id) DO UPDATE SET username=EXCLUDED.username
        RETURNING id`).Scan(&userID)
	assert.NoError(t, err)

	serviceID := 5
	targetDate := time.Now().AddDate(0, 0, 3).Truncate(24 * time.Hour)

	slotPending := "10:00:00"
	slotConfirmed := "11:00:00"

	_, err = db.Exec(`
        INSERT INTO bookings (user_id, service_id, booking_date, booking_time, guest_name, status)
        VALUES ($1, $2, $3, $4, 'Guest Pending', 'pending'),
               ($1, $2, $3, $5, 'Guest Confirmed', 'confirmed')`,
		userID, serviceID, targetDate, slotPending, slotConfirmed)
	assert.NoError(t, err)

	slots, err := repoBooking.GetAvailableSlots(context.Background(), serviceID, targetDate)

	assert.NoError(t, err)
	if assert.Len(t, slots, 1, "Should only return slots where status is not confirmed") {
		assert.Equal(t, "10:00", slots[0].Format("15:04"))
	}
}
