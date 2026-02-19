package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/models"
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
	return withMetrics(operation, func() error {

		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (telegram_id, username, first_name, last_name) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (telegram_id) 
			DO UPDATE SET username=EXCLUDED.username
			`, telegramID, userName, firstName, lastName)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	var operation = "get_user"

	return withMetricsValue(operation, func() (*models.User, error) {

		err := u.db.GetContext(ctx, &user, `
		SELECT *
		FROM users
		WHERE telegram_id=$1`, telegramID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		if err != nil {
			return nil, err
		}

		return &user, err
	})
}

func (u *UserRepo) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	var operation = "update_user_grade"
	return withMetrics(operation, func() error {

		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET grade=$1
		WHERE telegram_id=$2`,
			grade, telegramID)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err == nil && affected == 0 {
			//logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
			return models.ErrUserNotFound
		}
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	})
}

func (u *UserRepo) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	var isAdmin bool
	var operation = "check_is_admin"

	return withMetricsValue(operation, func() (bool, error) {

		err := u.db.QueryRowContext(ctx, `
	SELECT is_admin
	FROM users
	WHERE telegram_id=$1`,
			telegramID).Scan(&isAdmin)

		if errors.Is(err, sql.ErrNoRows) {
			//logger.Error("user_not_found", zap.Error(err), zap.String("operation", operation))
			return false, models.ErrUserNotFound
		}
		if err != nil {
			return false, err
		}

		return isAdmin, err
	})
}
