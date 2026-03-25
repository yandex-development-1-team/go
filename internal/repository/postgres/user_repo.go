package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

const getUserByEmailQuery = `
    SELECT id, telegram_nick, first_name, last_name, second_name, email,
		       phone_number, password_hash, role, status, invite_token, department,
					 position, manager_id, permissions, created_at, updated_at
    FROM staff
    WHERE email = $1`

const createUserQuery = `
	INSERT INTO staff(first_name, last_name, email, invite_token, password_hash)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, telegram_nick, first_name, email,
						role, status, permissions,
						created_at, updated_at
`

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) CreateStaff(ctx context.Context, userReq *models.UserAPI, hashPassword string) (*models.UserAPI, error) {
	var user dto.UserRow
	err := u.db.GetContext(ctx, &user, createUserQuery,
		userReq.Name,
		userReq.LastName,
		userReq.Email,
		userReq.InviteToken,
		hashPassword)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, models.ErrEmailAlreadyExist
		}
		return nil, err
	}

	return toUser(&user), nil
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error) {
	var user dto.UserRow
	err := u.db.GetContext(ctx, &user, getUserByEmailQuery, email)
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
}

func toUser(user *dto.UserRow) *models.UserAPI {
	return &models.UserAPI{
		ID:           user.ID,
		TelegramNick: derefString(user.TelegramNick),
		Name:         user.Name,
		LastName:     user.LastName,
		SecondName:   user.SecondName,
		Email:        user.Email,
		PhoneNumber:  derefString(user.PhoneNumber),
		Role:         user.Role,
		Status:       user.Status,
		Department:   derefString(user.Department),
		Position:     derefString(user.Position),
		ManagerID:    derefBigNumbers(user.ManagerID),
		InviteToken:  derefString(user.InviteToken),
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

func derefBigNumbers(n *int64) int64 {
	if n == nil {
		return 0
	}
	return *n
}
