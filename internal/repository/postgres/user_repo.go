package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/models"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error) {
	//operation := "get_user"
	//return withMetricsValue(operation, func() (*models.UserWithAuth, error) {

	query := `
    SELECT id, username, first_name, last_name, email, 
           role, status, permissions, password_hash, created_at, updated_at
    FROM users
    WHERE email = $1`

	var user userRow
	err := u.db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &models.UserWithAuth{
		User:     toUser(&user),
		PassHash: user.UserPass,
	}, nil
	//})
}

func toUser(user *userRow) *models.UserAPI {
	return &models.UserAPI{
		ID:           user.ID,
		TelegramNick: derefString(user.TelegramNick),
		Name:         strings.TrimSpace(derefString(user.FirstName) + " " + derefString(user.LastName)),
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		Permissions:  user.Permissions,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
