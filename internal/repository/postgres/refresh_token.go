package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	insertRefreshTokenQuery = `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)`

	getRefreshTokenForUpdateQuery = `
		SELECT id, user_id, token, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token = $1
		FOR UPDATE`

	revokeRefreshTokenQuery = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token = $1 AND revoked_at IS NULL`
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRTRequestTimeout     = errors.New("request timeout")
	ErrRTRequestCanceled    = errors.New("request canceled")
	ErrRTDatabase           = errors.New("database error")
)

type RefreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, rt *models.RefreshToken) error {
	const op = "create_refresh_token"
	_, err := r.db.ExecContext(ctx, insertRefreshTokenQuery, rt.UserID, rt.Token, rt.ExpiresAt)
	if err != nil {
		return r.checkError(op, err)
	}
	return nil
}

// CreateRefreshToken is a convenience method for the auth service interface.
func (r *RefreshTokenRepo) CreateRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	return r.Create(ctx, &models.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

func (r *RefreshTokenRepo) GetForUpdate(ctx context.Context, tx *sqlx.Tx, token string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := tx.GetContext(ctx, &rt, getRefreshTokenForUpdateQuery, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRefreshTokenNotFound
	}
	if err != nil {
		return nil, err
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
	res, err := r.db.ExecContext(ctx, revokeRefreshTokenQuery, token)
	if err != nil {
		return r.checkError(op, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return r.checkError(op, err)
	}
	if affected == 0 {
		return ErrRefreshTokenNotFound
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
