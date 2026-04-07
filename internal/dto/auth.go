package dto

import (
	"time"

	"github.com/lib/pq"
)

type LoginRequest struct {
	Login    string `json:"login"     binding:"required,email,max=255"`
	Password string `json:"password"  binding:"required,min=8,max=72"`
}

type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID           int64     `json:"id"`
	TelegramNick string    `json:"telegram_nick"`
	Name         string    `json:"name"`
	LastName     string    `json:"last_name"`
	SecondName   string    `json:"second_name"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Department   string    `json:"department"`
	Position     string    `json:"position"`
	ManagerID    int64     `json:"manager_id"`
	InviteToken  string    `json:"invite_token"`
	Permissions  []string  `json:"permissions"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRow struct {
	ID           int64          `db:"id"`
	TelegramNick *string        `db:"telegram_nick"`
	Name         string         `db:"first_name"`
	LastName     string         `db:"last_name"`
	SecondName   string         `db:"second_name"`
	Email        string         `db:"email"`
	PhoneNumber  *string        `db:"phone_number"`
	UserPass     string         `db:"password_hash"`
	Role         string         `db:"role"`
	Status       string         `db:"status"`
	InviteToken  *string        `db:"invite_token"`
	Department   *string        `db:"department"`
	Position     *string        `db:"position"`
	ManagerID    *int64         `db:"manager_id"`
	Permissions  pq.StringArray `db:"permissions"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterRequest struct {
	Name        string `json:"first_name"         binding:"required,min=2,max=255"`
	LastName    string `json:"last_name"          binding:"required,min=2,max=255"`
	Email       string `json:"email"              binding:"required,email,max=255"`
	Password    string `json:"password"           binding:"required,min=8,max=72"`
	InviteToken string `json:"invite_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
