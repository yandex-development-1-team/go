package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
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

// SpecialProjectRepository — CRUD спецпроектов.
type SpecialProjectRepository interface {
	Create(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error)
	GetByID(ctx context.Context, id int64) (*specialproject.DB, error)
	List(ctx context.Context, statusFilter *bool, searchQuery string, limit, offset int) ([]*specialproject.DB, int, error)
	Update(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error)
	Delete(ctx context.Context, id int64) error
}
