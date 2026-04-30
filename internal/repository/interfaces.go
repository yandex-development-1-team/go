package repository

import (
	"context"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

type StaffRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error)
	CreateStaff(ctx context.Context, userReq *models.UserAPI, hashPassword string) (*models.UserAPI, error)
	List(ctx context.Context, role, status, search string, limit, offset int) ([]dto.UserListItem, int, error)
	GetByID(ctx context.Context, id int64) (*dto.UserWithDetails, error)
	GetDashboard(ctx context.Context, managerId int64) (*dto.DashboardResponse, error)
	CreateStaffByAdmin(ctx context.Context, req *models.StaffAdminCreate) (*models.UserAPI, error)
	UpdateStaff(ctx context.Context, id int64, req *models.StaffAdminUpdate) (*models.UserAPI, error)
	UpdateStaffStatus(ctx context.Context, id int64, status string) (*models.UserAPI, error)
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
	GetBookingById(ctx context.Context, id int64) (*models.BookingAPI, error)
	GetBookingsList(ctx context.Context, filter *models.ApplicationFilter) (*models.BookingList, error)
	DeleteBooking(ctx context.Context, id int64) error
}

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, req *models.Application) error
	GetApplicationByID(ctx context.Context, id int64) (*models.Application, error)
	UpdateApplicationStatus(ctx context.Context, id int64, status string) error
	ApplicationsList(ctx context.Context, filter *models.ApplicationFilter) (*models.ApplicationList, error)
	DeleteApplication(ctx context.Context, id int64) error
}

//go:generate mockgen -source=interfaces.go -destination=../service/api/mocks/mock_boxlister.go -package=mocks
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
	GetSettings(ctx context.Context) (models.SettingsFormMessages, error)
	GetSettingsPermissions(ctx context.Context, role string) (models.SettingsPermissions, error)
	PutSettings(ctx context.Context, newSettings models.SettingsFormMessages) error
	PostSettings(ctx context.Context, newSettings models.SettingsPermissions) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error
	CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByStaffID(ctx context.Context, id int64) error
}

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
}

type SpecialProjectRepository interface {
	Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error)
	GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	List(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error)
	Update(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error)
	Delete(ctx context.Context, id int64) error
}

//go:generate mockgen -source=interfaces.go -destination=../service/api/mocks/mock_repository.go -package=mocks
type TxRepository interface {
	RunToTx(ctx context.Context, fn func(ctx context.Context) error) error
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type ResourcePageRepository interface {
	GetAll(ctx context.Context) ([]models.ResourcePage, error)
	GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error)
	Update(ctx context.Context, slug string, page models.ResourcePage) (*models.ResourcePage, error)
	Clear(ctx context.Context, slug string) (*models.ResourcePage, error)
	DeleteLink(ctx context.Context, slug string, id string) (*models.ResourcePage, error)
}

type Queryable interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
