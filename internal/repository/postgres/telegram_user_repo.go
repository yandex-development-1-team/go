package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
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

type TelegramUserRepo struct {
	db *sqlx.DB
}

func NewTelegramUserRepository(db *sqlx.DB) *TelegramUserRepo {
	return &TelegramUserRepo{db: db}
}

func (u *TelegramUserRepo) CreateUser(ctx context.Context, telegramID int64, userName, firstName, lastName string) error {
	const operation = "create_user"
	return repository.WithDBMetrics(operation, func() error {
		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		if _, err = tx.ExecContext(ctx, upsertUserQuery, telegramID, userName, firstName, lastName); err != nil {
			return err
		}
		return tx.Commit()
	})
}

func (u *TelegramUserRepo) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	const operation = "get_user"
	var user models.User

	return repository.WithDBMetricsValue(operation, func() (*models.User, error) {
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

func (u *TelegramUserRepo) UpdateUserGrade(ctx context.Context, telegramID int64, grade int) error {
	const operation = "update_user_grade"
	return repository.WithDBMetrics(operation, func() error {
		tx, err := u.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

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

func (u *TelegramUserRepo) IsAdmin(ctx context.Context, telegramID int64) (bool, error) {
	const operation = "check_is_admin"
	var isAdmin bool

	return repository.WithDBMetricsValue(operation, func() (bool, error) {
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
