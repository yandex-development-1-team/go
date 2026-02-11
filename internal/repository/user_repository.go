package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrUserNotFound    = errors.New("user not found")
	ErrRequestCanceled = errors.New("request canceled")
	ErrDatabase        = errors.New("database error")
)

type UserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
}

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error {
	var operation = "create_user"

	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return u.checkError(operation, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
	INSERT INTO users (telegram_id, username, first_name, last_name) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (telegram_id) 
	DO UPDATE SET username=EXCLUDED.username
	`, telegramID, userName, firstName, lastName)
	if err != nil {
		return u.checkError(operation, err)
	}

	err = tx.Commit()
	if err != nil {
		return u.checkError(operation, err)
	}

	return nil
}

func (u *UserRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	var operation = "get_user"

	err := u.db.GetContext(ctx, &user, `
		SELECT *
		FROM users
		WHERE telegram_id=$1`, telegramID)
	if errors.Is(err, sql.ErrNoRows) {
		logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, u.checkError(operation, err)
	}

	return &user, err
}

func (u *UserRepo) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	var operation = "update_user_grade"

	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return u.checkError(operation, err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET grade=$1
		WHERE telegram_id=$2`,
		grade, telegramID)
	if err != nil {
		return u.checkError(operation, err)
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return ErrUserNotFound
	}
	if err != nil {
		return u.checkError(operation, err)
	}

	err = tx.Commit()
	if err != nil {
		return u.checkError(operation, err)
	}

	return nil
}

func (u *UserRepo) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	var isAdmin bool
	var operation = "check_is_admin"

	err := u.db.QueryRowContext(ctx, `
	SELECT is_admin
	FROM users
	WHERE telegram_id=$1`,
		telegramID).Scan(&isAdmin)

	if errors.Is(err, sql.ErrNoRows) {
		logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, u.checkError(operation, err)
	}

	return isAdmin, err
}

func (u *UserRepo) checkError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		logger.Error("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return ErrRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Error("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return ErrRequestTimeout
	}

	logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return ErrDatabase
}
