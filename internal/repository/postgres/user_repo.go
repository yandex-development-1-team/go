package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yandex-development-1-team/go/internal/models"
)

type userRow struct {
	ID           int64   `db:"id"`
	TelegramNick *string `db:"telegram_nick"`
	Name         string  `db:"name"`
	Email        string  `db:"email"`
	UserPass     string  `db:"password_hash"`
	Role         string  `db:"role"`
	Status       string  `db:"status"`
	// InviteToken *string        `db:"invite_token"`
	Permissions pq.StringArray `db:"permissions"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

const getUserByEmailQuery = `
    SELECT id, telegram_nick, name, email,
           role, status, permissions, password_hash, created_at, updated_at
    FROM staff
    WHERE email = $1`

const createUserQuery = `
	INSERT INTO staff(name, email, role, invite_token, password_hash)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, telegram_nick, name, email,
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
	var user userRow
	err := u.db.GetContext(ctx, &user, createUserQuery,
		userReq.Name,
		userReq.Email,
		// userReq.Password,
		userReq.Role,
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
	var user userRow
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

func toUser(user *userRow) *models.UserAPI {
	return &models.UserAPI{
		ID:           user.ID,
		TelegramNick: derefString(user.TelegramNick),
		// Name:         strings.TrimSpace(derefString(user.FirstName) + " " + derefString(user.LastName)),
		Name:        user.Name,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		Permissions: user.Permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
