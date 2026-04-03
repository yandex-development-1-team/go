package repository

import (
	"context"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

type StaffRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error)
	CreateStaff(ctx context.Context, userReq *models.UserAPI, hashPassword string) (*models.UserAPI, error)
}

type TelegramUserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
}

type BookingRepository interface {
	CreateBooking(ctx context.Context, b *models.Booking) (int64, error)
	GetAvailableSlots(ctx context.Context, serviceID int, date time.Time) ([]time.Time, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]models.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error
}

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error)
	GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error)
	GetApplicationByID(ctx context.Context, id int64) (*models.Application, error)
	UpdateApplication(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error)
	DeleteApplication(ctx context.Context, id int64) error
}

//go:generate mockgen -source=../../repository/postgres/interfaces.go -destination=mocks/mock_boxlister.go -package=mocks
type BoxSolutionRepository interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
	GetServiceByID(ctx context.Context, serviceID int64) (*models.Service, error)
	CreateBox(ctx context.Context, box *models.BoxCreate) (*models.Service, error)
	GetAvailableSlotsByServiceID(ctx context.Context, serviceID int64) ([]models.BoxAvailableSlot, error)
	CheckSlotAvailability(ctx context.Context, serviceID int64, slot models.BoxAvailableSlot) (bool, error)
	UpdateService(ctx context.Context, id int64, service *models.BoxUpdate) error
	SoftDeleteService(ctx context.Context, serviceID int64) error
	UpdateServiceStatus(ctx context.Context, serviceID int64, status models.ServiceStatus) (*models.BoxUpdateStatusResult, error)
	UpdateServiceSlots(ctx context.Context, id int64, slots *models.BoxNewSlots) error
	DeleteServiceSlots(ctx context.Context, id int64) error
	GetServicesByStatus(ctx context.Context, status *models.ServiceStatus) ([]models.Service, error)
	List(ctx context.Context, query models.BoxList) (*models.BoxListResult, error)
}

type SessionRepository interface {
	SaveSession(ctx context.Context, userID int64, state string, data map[string]interface{}) error
	GetSession(ctx context.Context, userID int64) (*models.UserSession, error)
	ClearSession(ctx context.Context, userID int64) error
	UpdateSessionState(ctx context.Context, userID int64, newState string) error
}

type SettingsRepository interface {
	GetSettings(ctx context.Context) ([]models.SettingRow, error)
	PutSettings(ctx context.Context, newSettings []models.Setting) (time.Time, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error
	CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
}

type SpecialProjectRepository interface {
	Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error)
	GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	List(ctx context.Context, statusFilter *bool, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error)
	Update(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error)
	Delete(ctx context.Context, id int64) error
}

type TxRepository interface {
	RunToTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type ResourcePageRepository interface {
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
	GetBySlugTx(ctx context.Context, queryable Queryable, slug string, lockForUpdate bool) (*models.ResourcePage, error)
	UpdatePageContentAndLinksTx(ctx context.Context, tx *sqlx.Tx, slug string, title string, content string, links []models.ResourcePageLink) error
	GetAllSummaries(ctx context.Context) ([]*models.ResourcePage, error)
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type Queryable interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
