package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/resourcepage"
	"github.com/yandex-development-1-team/go/internal/specialproject"
)

// UserRepository — доступ к пользователям (например по email для логина).
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error)
}

// SettingsRepository — чтение настроек из хранилища.
type SettingsRepository interface {
	GetSettings(ctx context.Context) ([]models.Setting, error)
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
	Create(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error)
	GetByID(ctx context.Context, id int64) (*specialproject.DB, error)
	List(ctx context.Context, statusFilter *bool, searchQuery string, limit, offset int) ([]*specialproject.DB, int, error)
	Update(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error)
	Delete(ctx context.Context, id int64) error
}

type ResourcePageRepository interface {
	// GetBySlug возвращает страницу по slug без транзакции.
	GetBySlug(ctx context.Context, slug string) (*resourcepage.ResourcePage, error)
	// GetBySlugTx возвращает страницу по slug. Может работать внутри транзакции.
	// lockForUpdate указывает, нужно ли блокировать строку для обновления.
	GetBySlugTx(ctx context.Context, queryable Queryable, slug string, lockForUpdate bool) (*resourcepage.ResourcePage, error)

	// UpdatePageContentAndLinksTx обновляет title, content и links ВНУТРИ переданной транзакции.
	UpdatePageContentAndLinksTx(ctx context.Context, tx *sqlx.Tx, slug string, title string, content string, links []resourcepage.Link) error

	// GetAllSummaries возвращает краткую информацию о всех страницах без транзакции.
	GetAllSummaries(ctx context.Context) ([]*resourcepage.ResourcePage, error)

	// BeginTx начинает новую транзакцию.
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// Implements sqlx.DB and sqlx.Tx
type Queryable interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
