package models

import (
	"errors"
	"time"
)

const (
	PermBookingsView      = "bookings:view"
	PermBookingsEdit      = "bookings:edit"
	PermBookingsDelete    = "bookings:delete"
	PermBoxesCreate       = "boxes:create"
	PermBoxesEdit         = "boxes:edit"
	PermBoxesDelete       = "boxes:delete"
	PermSpecProjectView   = "specproject:view"
	PermSpecPrijectEdit   = "specproject:edit"
	PermSpecProjectDelete = "specproject:delete"
	PermAnalyticsView     = "analytics:view"
	PermAnalyticsDownload = "analytics:download"
	PermEvents            = "events:yes"
	PermAboutUs           = "aboutus:yes"
	PermFAQ               = "faq:yes"
)

var (
	ErrRequestTimeout         = errors.New("request timeout")
	ErrRequestCanceled        = errors.New("request canceled")
	ErrDatabase               = errors.New("database error")
	ErrSlotOccupied           = errors.New("slot is already occupied")
	ErrInvalidInput           = errors.New("invalid input data")
	ErrBookingNotFound        = errors.New("booking not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrSpecialProjectNotFound = errors.New("special project not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserBlocked            = errors.New("user blocked")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrValidation             = errors.New("validation error")
	ErrForbidden              = errors.New("forbidden")
	ErrCache                  = errors.New("cache error")
	ErrEmailAlreadyExist      = errors.New("user already exist")
	ErrSessionNotFound        = errors.New("session not found")
	ErrApplicationNotFound    = errors.New("application not found")
	ErrTokenNotFound          = errors.New("token not found")
	ErrBoxSolutionNotFound    = errors.New("box solution not found")
	ErrSlotsNotFound          = errors.New("slots not found")
)

type ApplicationURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type RefreshToken struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Role      string    `db:"role"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type PasswordResetToken struct {
	ID        int64      `db:"id"`
	UserID    int64      `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

type User struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	Grade      int       `db:"grade"`
	IsAdmin    bool      `db:"is_admin"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// UserAPI is the API/domain representation of a user (auth and handlers).
type UserAPI struct {
	ID           int64
	TelegramNick string
	Name         string
	LastName     string
	SecondName   string
	Email        string
	PhoneNumber  string
	Role         string
	Status       string
	Department   string
	Position     string
	InviteToken  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserWithAuth holds user and password hash for auth flow.
type UserWithAuth struct {
	User     *UserAPI `json:"user"`
	PassHash string   `json:"-"`
}

// AuthResult is returned on successful login.
type AuthResult struct {
	User         *UserAPI `json:"user"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
}



type UserSession struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	CurrentState string                 `json:"current_state"`
	StateData    map[string]interface{} `json:"state_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type RefreshResponse struct {
	Token        string
	RefreshToken string
}
