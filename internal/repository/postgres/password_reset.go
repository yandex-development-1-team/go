package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

type PasswordResetRepository struct {
	db *sqlx.DB
}

func NewPasswordResetRepository(db *sqlx.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	return err
}

func (r *PasswordResetRepository) GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var prt models.PasswordResetToken
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`
	err := r.db.GetContext(ctx, &prt, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrTokenNotFound
		}
		return nil, err
	}
	return &prt, nil
}

