package repository

import (
	"context"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

// StaffRepository — доступ к сотрудникам (таблица staff, логин по email).
type StaffRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error)
	CreateStaff(ctx context.Context, userReq *models.UserAPI, hashPassword string) (*models.UserAPI, error)
}

// TelegramUserRepository — пользователи бота (таблица users, telegram_id).
type TelegramUserRepository interface {
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
	GetApplicationByID(ctx context.Context, id int64) (*models.Application, error)
	UpdateApplication(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error)
	DeleteApplication(ctx context.Context, id int64) error
}

// BoxSolutionRepository — коробочные решения (services).
type BoxSolutionRepository interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
	GetServiceByID(ctx context.Context, serviceID int) (models.Service, error)
	GetAvailableSlotsByServiceID(ctx context.Context, serviceID int) ([]models.AvailableSlot, error)
	GetAvailableTimeSlotsByDate(ctx context.Context, serviceID int, date string) ([]string, error)
}

// SessionRepository — сессии пользователей (Redis).
type SessionRepository interface {
	SaveSession(ctx context.Context, userID int64, state string, data map[string]interface{}) error
	GetSession(ctx context.Context, userID int64) (*models.UserSession, error)
	ClearSession(ctx context.Context, userID int64) error
	UpdateSessionState(ctx context.Context, userID int64, newState string) error
}

// SettingsRepository — чтение настроек из хранилища.
type SettingsRepository interface {
	GetSettings(ctx context.Context) ([]models.SettingRow, error)
	PutSettings(ctx context.Context, newSettings []models.Setting) (time.Time, error)
}

// RefreshTokenRepository — хранение и инвалидация refresh-токенов.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error
	CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
}

// SpecialProjectRepository — CRUD for spesial projects.
type SpecialProjectRepository interface {
	Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error)
	GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	List(ctx context.Context, statusFilter *bool, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error)
	Update(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error)
	Delete(ctx context.Context, id int64) error
}

// TxRepository — атомарность работы с бд.
type TxRepository interface {
	RunToTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ResourcePageRepository — страницы ресурсов.
type ResourcePageRepository interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
	GetBySlugTx(ctx context.Context, queryable Queryable, slug string, lockForUpdate bool) (*models.ResourcePage, error)
	UpdatePageContentAndLinksTx(ctx context.Context, tx *sqlx.Tx, slug string, title string, content string, links []models.ResourcePageLink) error
	GetAllSummaries(ctx context.Context) ([]*models.ResourcePage, error)
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// Queryable — общее для sqlx.DB и sqlx.Tx.
type Queryable interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
