package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type RefreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{
		db: db,
	}
}

func (rf *RefreshTokenRepo) CreateRefreshToken(ctx context.Context, userId int64, token string, expiresAt time.Time) error {
	query := `
	INSERT INTO refresh_tokens (user_id, token, expires_at)
	VALUES ($1, $2, $3)
	`
	tx, err := rf.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, query, userId, token, expiresAt)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
