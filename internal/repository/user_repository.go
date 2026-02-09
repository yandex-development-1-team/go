package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

var (
	ErrRequestTimeout  = errors.New("request timeout")
	ErrUserNotFound    = errors.New("user not found")
	ErrRequestCanceled = errors.New("request canceles")
	ErrDatabase        = errors.New("database error")
)

type UserRepository interface {
	CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) (*models.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error
	IsAdmin(ctx context.Context, telegramID int64) (bool, error)
}

type UserRepo struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserRepository(db *sqlx.DB, logger *zap.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

func (u *UserRepo) CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) (*models.User, error) {
	var user models.User
	var operation = "create_user"

	err := u.db.GetContext(ctx, &user, `
	INSERT INTO users (telegram_id, username, first_name, last_name)
	SELECT $1, $2, $3, $4
	WHERE NOT EXISTS (
			SELECT 1 FROM users WHERE telegram_id = $1
	)
	RETURNING *`,
		telegramID, userName, firstName, lastName)
	if errors.Is(err, sql.ErrNoRows) {
		u.logger.Error("user_already_exist", zap.Error(models.ErrAlreadyExist), zap.String("operation", operation))
		return &user, models.ErrAlreadyExist
	}
	if err != nil {
		return &user, u.checkError(operation, err)
	}

	return &user, err
}

func (u *UserRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	var operation = "get_user"

	err := u.db.GetContext(ctx, &user, `
		SELECT *
		FROM users
		WHERE telegram_id=$1`, telegramID)
	if err == sql.ErrNoRows {
		u.logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return &user, ErrUserNotFound
	}
	if err != nil {
		return &user, u.checkError(operation, err)
	}

	return &user, err
}

func (u *UserRepo) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	var operation = "update_user_grade"

	result, err := u.db.ExecContext(ctx, `
		UPDATE users
		SET grade=$1
		WHERE telegram_id=$2`,
		grade, telegramID)

	if err != nil {
		return u.checkError(operation, err)
	}

	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		u.logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return ErrUserNotFound
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
		u.logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
		return isAdmin, ErrUserNotFound
	}
	if err != nil {
		return isAdmin, u.checkError(operation, err)
	}

	return isAdmin, err
}

func (u *UserRepo) checkError(operation string, err error) error {
	if errors.Is(err, context.Canceled) {
		u.logger.Error("canceled_by_context", zap.Error(err), zap.String("operation", operation))
		return ErrRequestCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		u.logger.Error("canceled_by_timeout", zap.Error(err), zap.String("operation", operation))
		return ErrRequestTimeout
	}

	u.logger.Error("database_error", zap.Error(err), zap.String("operation", operation))
	return ErrDatabase
}
