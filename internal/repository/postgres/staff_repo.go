package postgres

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
const insertStaffAdminQuery = `
	INSERT INTO staff (
		telegram_nick, first_name, last_name, email,
		password_hash, role, status, invite_token, permissions)
	VALUES ($1, $2, $3, $4, '', $5::user_role_type, $6::user_status_type, $7, $8)
	RETURNING id, telegram_nick, first_name, last_name, second_name, email, phone_number,
		role, status, invite_token, department, position, manager_id, permissions,
		created_at, updated_at`

const updateStaffQuery = `
	UPDATE staff SET
		first_name = COALESCE($2, first_name),
		email = COALESCE($3, email),
		role = COALESCE($4::user_role_type, role),
		status = COALESCE($5::user_status_type, status),
		permissions = COALESCE($6, permissions),
		telegram_nick = COALESCE($7, telegram_nick)
	WHERE id = $1
	RETURNING id, telegram_nick, first_name, last_name, second_name, email, phone_number,
		role, status, invite_token, department, position, manager_id, permissions,
		created_at, updated_at`

const blockStaffQuery = `
	UPDATE staff SET status = 'blocked'::user_status_type
	WHERE id = $1
	RETURNING id, telegram_nick, first_name, last_name, second_name, email, phone_number,
		role, status, invite_token, department, position, manager_id, permissions,
		created_at, updated_at`

type StaffRepo struct {
	db *sqlx.DB
}

func NewStaffRepo(db *sqlx.DB) *StaffRepo {
	return &StaffRepo{
		db: db,
	}
}

func (u *StaffRepo) CreateStaff(ctx context.Context, userReq *models.UserAPI, hashPassword string) (*models.UserAPI, error) {
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

func (u *StaffRepo) GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error) {
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

func (u *StaffRepo) CreateStaffByAdmin(ctx context.Context, req *models.StaffAdminCreate) (*models.UserAPI, error) {
	var row dto.UserRow

	var tg sql.NullString
	if req.TelegramNick != nil && *req.TelegramNick != "" {
		tg = sql.NullString{String: *req.TelegramNick, Valid: true}
	}

	err := u.db.GetContext(ctx, &row, insertStaffAdminQuery,
		tg,
		req.Name,
		"",
		req.Email,
		req.Role,
		req.Status,
		req.InviteToken,
		pq.Array(req.Permissions),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, models.ErrEmailAlreadyExist
		}
		return nil, err
	}
	return toUser(&row), nil
}

func (u *StaffRepo) UpdateStaff(ctx context.Context, id int64, req *models.StaffAdminUpdate) (*models.UserAPI, error) {
	var row dto.UserRow

	var tg interface{}
	if req.TelegramNick != nil {
		if *req.TelegramNick == "" {
			tg = nil
		} else {
			tg = *req.TelegramNick
		}
	}

	var perms interface{}
	if req.Permissions != nil {
		perms = pq.Array(*req.Permissions)
	}

	err := u.db.GetContext(ctx, &row, updateStaffQuery,
		id,
		req.Name,
		req.Email,
		req.Role,
		req.Status,
		perms,
		tg,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUser(&row), nil
}

func (u *StaffRepo) BlockStaff(ctx context.Context, id int64) (*models.UserAPI, error) {
	var row dto.UserRow
	err := u.db.GetContext(ctx, &row, blockStaffQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUser(&row), nil
}
