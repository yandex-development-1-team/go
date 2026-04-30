package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
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
	_, err := r.getDB(ctx).ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	return err
}

func (r *PasswordResetRepository) GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var prt models.PasswordResetToken
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
		FOR UPDATE
	`
	err := sqlx.GetContext(ctx, r.getDB(ctx), &prt, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTokenNotFound
		}
		return nil, err
	}
	return &prt, nil
}

func (r *PasswordResetRepository) DeleteToken(ctx context.Context, id int64) error {
	query := `
		UPDATE password_reset_tokens 
		SET used_at = NOW() 
		WHERE id = $1 AND used_at IS NULL
		`
	resp, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	affected, err := resp.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return models.ErrInvalidCredentials
	}
	return nil
}

func (r *PasswordResetRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctxutil.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
