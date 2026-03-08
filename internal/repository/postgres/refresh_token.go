package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRTRequestTimeout     = errors.New("request timeout")
	ErrRTRequestCanceled    = errors.New("request canceled")
	ErrRTDatabase           = errors.New("database error")
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *models.RefreshToken) error

	GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
}

type RefreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, rt *models.RefreshToken) error {
	const op = "create_refresh_token"

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, rt.UserID, rt.Token, rt.ExpiresAt)
	if err != nil {
		return r.checkError(op, err)
	}
	return nil
}

func (RefreshTokenRepo) GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error) {
	const op = "get_refresh_token"

	var rt models.RefreshToken
	err := tx.GetContext(ctx, &rt, `
		SELECT id, user_id, token, expires_at, revoked_at, created_at 
		FROM refresh_tokens 
		WHERE token=$1
		FOR UPDATE
		`, token)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, ErrRefreshTokenRevoked
	}

	now := time.Now().UTC()
	if rt.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}
	if !rt.ExpiresAt.After(now) {
		return nil, ErrRefreshTokenExpired
	}

	return &rt, nil
}

func (r *RefreshTokenRepo) Revoke(ctx context.Context, token string) error {
	const op = "revoke_refresh_token"

	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at= NOW()
		WHERE token=$1 AND revoked_at IS NULL
	`, token)
	if err != nil {
		return r.checkError(op, err)
	}
	return nil
}

func (r *RefreshTokenRepo) checkError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		logger.Error("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return ErrRTRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Error("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return ErrRTRequestTimeout
	}

	logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return ErrRTDatabase
}
