package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	upsertUserQuery = `
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id)
		DO UPDATE SET username = EXCLUDED.username`

	getUserByTelegramIDQuery = `
		SELECT id, telegram_id, username, first_name, last_name, grade, is_admin, created_at, updated_at
		FROM users WHERE telegram_id = $1`

	updateUserGradeQuery = `
		UPDATE users SET grade = $1 WHERE telegram_id = $2`

	isAdminQuery = `
		SELECT is_admin FROM users WHERE telegram_id = $1`
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (u *UserRepo) CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error {
	const operation = "create_user"
	return withMetrics(operation, func() error {
		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if _, err = tx.ExecContext(ctx, upsertUserQuery, telegramID, userName, firstName, lastName); err != nil {
			return err
		}
		return tx.Commit()
	})
}

func (u *UserRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	const operation = "get_user"
	var user models.User

	return withMetricsValue(operation, func() (*models.User, error) {
		err := u.db.GetContext(ctx, &user, getUserByTelegramIDQuery, telegramID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		if err != nil {
			return nil, err
		}
		return &user, nil
	})
}

func (u *UserRepo) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	const operation = "update_user_grade"
	return withMetrics(operation, func() error {
		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(ctx, updateUserGradeQuery, grade, telegramID)
		if err != nil {
			return err
		}

		affected, err := result.RowsAffected()
		if err == nil && affected == 0 {
			return models.ErrUserNotFound
		}
		if err != nil {
			return err
		}
		return tx.Commit()
	})
}

func (u *UserRepo) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	const operation = "check_is_admin"
	var isAdmin bool

	return withMetricsValue(operation, func() (bool, error) {
		err := u.db.QueryRowContext(ctx, isAdminQuery, telegramID).Scan(&isAdmin)
		if errors.Is(err, sql.ErrNoRows) {
			return false, models.ErrUserNotFound
		}
		if err != nil {
			return false, err
		}
		return isAdmin, nil
	})
}
