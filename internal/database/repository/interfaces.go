package repository

import (
	"context"
	"time"

	"github.com/yandex-development-1-team/go/internal/models"
)

// UserRepository — пользователи (бот, telegram_id).
type UserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
}

// BookingRepository — бронирования.
type BookingRepository interface {
	CreateBooking(ctx context.Context, b *models.Booking) (int64, error)
	GetAvailableSlots(ctx context.Context, serviceID int, date time.Time) ([]time.Time, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error
}

// ApplicationRepository — заявки.
type ApplicationRepository interface {
	CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error)
	GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error)
}

// BoxSolutionRepository — коробочные решения (services).
type BoxSolutionRepository interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
	GetServiceByID(ctx context.Context, serviceID int) (models.Service, error)
}

// SessionRepository — сессии пользователей (Redis).
type SessionRepository interface {
	SaveSession(ctx context.Context, userID int64, state string, data map[string]interface{}) error
	GetSession(ctx context.Context, userID int64) (*models.UserSession, error)
	ClearSession(ctx context.Context, userID int64) error
	UpdateSessionState(ctx context.Context, userID int64, newState string) error
}
