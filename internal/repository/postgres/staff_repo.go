package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

const getUserByEmailQuery = `
    SELECT id, telegram_nick, first_name, last_name, second_name, email,
           phone_number, password_hash, role, status, invite_token, department,
           position, supervisor, address, created_at, updated_at
    FROM staff
    WHERE email = $1`

const createUserQuery = `
	INSERT INTO staff(first_name, last_name, email, invite_token, password_hash)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, telegram_nick, first_name, email,
						role, status, created_at, updated_at
`

const insertStaffAdminQuery = `
    INSERT INTO staff (
        first_name, last_name, second_name, email,
        phone_number, password_hash, role, status,
        department, position, invite_token, supervisor, address)
    VALUES ($1, $2, $3, $4, $5, '', $6::user_role_type, $7::user_status_type,
        $8, $9, $10, $11, $12)
    RETURNING id, telegram_nick, first_name, last_name, second_name, email,
        phone_number, role, status, invite_token, department, position,
        supervisor, address, created_at, updated_at`

const updateStaffQuery = `
    UPDATE staff SET
        first_name   = COALESCE($2, first_name),
        last_name    = COALESCE($3, last_name),
        second_name  = COALESCE($4, second_name),
        email        = COALESCE($5, email),
        role         = COALESCE($6::user_role_type, role),
        status       = COALESCE($7::user_status_type, status),
        phone_number = COALESCE($8, phone_number),
        department   = COALESCE($9, department),
        position     = COALESCE($10, position),
        supervisor   = COALESCE($11, supervisor),
        address      = COALESCE($12, address)
    WHERE id = $1
    RETURNING id, telegram_nick, first_name, last_name, second_name, email,
        phone_number, role, status, invite_token, department, position,
        supervisor, address, created_at, updated_at`

const blockStaffQuery = `
    UPDATE staff SET status = 'blocked'::user_status_type
    WHERE id = $1
    RETURNING id, telegram_nick, first_name, last_name, second_name, email,
        phone_number, role, status, invite_token, department, position,
        supervisor, address, created_at, updated_at`

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
		InviteToken:  derefString(user.InviteToken),
		Supervisor:   derefString(user.Supervisor),
		Address:      derefString(user.Address),
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

func (u *StaffRepo) List(ctx context.Context, role, status, search string, limit, offset int) ([]dto.UserListItem, int, error) {
	query := `
		SELECT id, telegram_nick, first_name, last_name, role, status, created_at
		FROM staff
		WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM staff WHERE 1=1`

	args := []any{}
	argPos := 1

	if role != "" {
		roles := strings.Split(role, ",")
		query += fmt.Sprintf(" AND role = ANY($%d)", argPos)
		countQuery += fmt.Sprintf(" AND role = ANY($%d)", argPos)
		args = append(args, pq.Array(roles))
		argPos++
	}

	if status != "" {
		statuses := strings.Split(status, ",")
		query += fmt.Sprintf(" AND status = ANY($%d)", argPos)
		countQuery += fmt.Sprintf(" AND status = ANY($%d)", argPos)
		args = append(args, pq.Array(statuses))
		argPos++
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query += fmt.Sprintf(" AND (telegram_nick ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos, argPos)
		countQuery += fmt.Sprintf(" AND (telegram_nick ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argPos, argPos, argPos, argPos)
		args = append(args, searchPattern)
		argPos++
	}

	var total int
	if err := u.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	type userRow struct {
		ID           int64        `db:"id"`
		TelegramNick *string      `db:"telegram_nick"`
		FirstName    string       `db:"first_name"`
		LastName     string       `db:"last_name"`
		Role         string       `db:"role"`
		Status       string       `db:"status"`
		CreatedAt    sql.NullTime `db:"created_at"`
	}

	var rows []userRow
	if err := u.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}

	items := make([]dto.UserListItem, 0, len(rows))
	for _, row := range rows {
		name := row.FirstName
		if row.LastName != "" {
			name += " " + row.LastName
		}
		items = append(items, dto.UserListItem{
			ID:           row.ID,
			TelegramNick: derefString(row.TelegramNick),
			Name:         name,
			Role:         row.Role,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt.Time,
		})
	}

	return items, total, nil
}

func (u *StaffRepo) GetByID(ctx context.Context, id int64) (*dto.UserWithDetails, error) {
	userQuery := `
    SELECT id, telegram_nick, first_name, last_name, second_name, email,
           phone_number, role, status, department, position,
           supervisor, address, created_at, updated_at
    FROM staff
    WHERE id = $1`

	type userRow struct {
		ID           int64        `db:"id"`
		TelegramNick *string      `db:"telegram_nick"`
		FirstName    string       `db:"first_name"`
		LastName     string       `db:"last_name"`
		SecondName   string       `db:"second_name"`
		Email        string       `db:"email"`
		PhoneNumber  *string      `db:"phone_number"`
		Role         string       `db:"role"`
		Status       string       `db:"status"`
		Department   *string      `db:"department"`
		Position     *string      `db:"position"`
		Supervisor   *string      `db:"supervisor"`
		Address      *string      `db:"address"`
		CreatedAt    sql.NullTime `db:"created_at"`
		UpdatedAt    sql.NullTime `db:"updated_at"`
	}

	var user userRow
	if err := u.db.GetContext(ctx, &user, userQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}

	bookingsQuery := `
		SELECT b.id, s.name as service_name, b.booking_date, b.status, b.created_at
		FROM bookings b
		JOIN services s ON b.service_id = s.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC`

	type bookingRow struct {
		ID          int64        `db:"id"`
		ServiceName string       `db:"service_name"`
		BookingDate sql.NullTime `db:"booking_date"`
		Status      string       `db:"status"`
		CreatedAt   sql.NullTime `db:"created_at"`
	}

	var bookingRows []bookingRow
	if err := u.db.SelectContext(ctx, &bookingRows, bookingsQuery, id); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	bookings := make([]dto.UserBookingItem, 0, len(bookingRows))
	for _, b := range bookingRows {
		bookings = append(bookings, dto.UserBookingItem{
			ID:          b.ID,
			ServiceName: b.ServiceName,
			BookingDate: b.BookingDate.Time,
			Status:      b.Status,
			CreatedAt:   b.CreatedAt.Time,
		})
	}

	visitHistoryQuery := `
		SELECT s.name as box_name, b.booking_date as visited_at
		FROM bookings b
		JOIN services s ON b.service_id = s.id
		WHERE b.user_id = $1 AND b.status = 'confirmed'
		ORDER BY b.booking_date DESC`

	type visitRow struct {
		BoxName   string       `db:"box_name"`
		VisitedAt sql.NullTime `db:"visited_at"`
	}

	var visitRows []visitRow
	if err := u.db.SelectContext(ctx, &visitRows, visitHistoryQuery, id); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	visitHistory := make([]dto.VisitHistoryItem, 0, len(visitRows))
	for _, v := range visitRows {
		visitHistory = append(visitHistory, dto.VisitHistoryItem{
			BoxName:   v.BoxName,
			VisitedAt: v.VisitedAt.Time,
		})
	}

	favoritesQuery := `
		SELECT service_id
		FROM user_favorites
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var favoriteBoxes []int64
	if err := u.db.SelectContext(ctx, &favoriteBoxes, favoritesQuery, id); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if favoriteBoxes == nil {
		favoriteBoxes = []int64{}
	}
	if bookings == nil {
		bookings = []dto.UserBookingItem{}
	}
	if visitHistory == nil {
		visitHistory = []dto.VisitHistoryItem{}
	}

	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}

	return &dto.UserWithDetails{
		ID:            user.ID,
		TelegramNick:  derefString(user.TelegramNick),
		Name:          name,
		LastName:      user.LastName,
		SecondName:    user.SecondName,
		Email:         user.Email,
		PhoneNumber:   derefString(user.PhoneNumber),
		Role:          user.Role,
		Status:        user.Status,
		Department:    derefString(user.Department),
		Position:      derefString(user.Position),
		Supervisor:    derefString(user.Supervisor),
		Address:       derefString(user.Address),
		CreatedAt:     user.CreatedAt.Time,
		UpdatedAt:     user.UpdatedAt.Time,
		Bookings:      bookings,
		VisitHistory:  visitHistory,
		FavoriteBoxes: favoriteBoxes,
	}, nil
}

func (u *StaffRepo) CreateStaffByAdmin(ctx context.Context, req *models.StaffAdminCreate) (*models.UserAPI, error) {
	var row dto.UserRow

	err := u.db.GetContext(ctx, &row, insertStaffAdminQuery,
		req.FirstName,
		req.LastName,
		req.SecondName,
		req.Email,
		req.PhoneNumber,
		req.Role,
		req.Status,
		req.Department,
		req.Position,
		req.InviteToken,
		req.Supervisor,
		req.Address,
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

	err := u.db.GetContext(ctx, &row, updateStaffQuery,
		id,
		req.FirstName,
		req.LastName,
		req.SecondName,
		req.Email,
		req.Role,
		req.Status,
		req.PhoneNumber,
		req.Department,
		req.Position,
		req.Supervisor,
		req.Address,
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
