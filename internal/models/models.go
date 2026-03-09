package models

import (
	"errors"
	"time"
)

var (
	ErrRequestTimeout     = errors.New("request timeout")
	ErrRequestCanceled    = errors.New("request canceled")
	ErrDatabase           = errors.New("database error")
	ErrCache              = errors.New("cache error")
	ErrSlotOccupied       = errors.New("slot is already occupied")
	ErrInvalidInput       = errors.New("invalid input data")
	ErrBookingNotFound    = errors.New("booking not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBlocked        = errors.New("user blocked")
)

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
type Booking struct {
	ID                int64      `db:"id"`
	UserID            int64      `db:"user_id"`
	ServiceID         int16      `db:"service_id"`
	BookingDate       time.Time  `db:"booking_date"`
	BookingTime       *time.Time `db:"booking_time"`
	GuestName         string     `db:"guest_name"`
	GuestOrganization string     `db:"guest_organization"`
	GuestPosition     string     `db:"guest_position"`
	VisitType         string     `db:"visit_type"`
	Status            string     `db:"status"`
	TrackerTicketID   string     `db:"tracker_ticket_id"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

// UserAPI — пользователь для API (без пароля).
type UserAPI struct {
	ID           int64     `json:"id"`
	TelegramNick string    `json:"telegram_nick"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Permissions  []string  `json:"permissions"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserWithAuth — пользователь с хешем пароля (репозиторий → сервис).
type UserWithAuth struct {
	User     *UserAPI
	PassHash string
}

// AuthResult — результат успешного логина.
type AuthResult struct {
	User         *UserAPI
	Token        string
	RefreshToken string
}

type UserSession struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	CurrentState string                 `json:"current_state"`
	StateData    map[string]interface{} `json:"state_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type ApplicationType string
type ApplicationSource string
type ApplicationStatus string

const (
	ApplicationTypeBox            ApplicationType = "box"
	ApplicationTypeSpecialProject ApplicationType = "special_project"

	ApplicationSourceTelegramBot ApplicationSource = "telegram_bot"
	ApplicationSourceManual      ApplicationSource = "manual"

	ApplicationStatusQueue      ApplicationStatus = "queue"
	ApplicationStatusInProgress ApplicationStatus = "in_progress"
	ApplicationStatusDone       ApplicationStatus = "done"
)

func (t ApplicationType) Valid() bool {
	switch t {
	case ApplicationTypeBox, ApplicationTypeSpecialProject:
		return true
	}
	return false
}

func (s ApplicationSource) Valid() bool {
	switch s {
	case ApplicationSourceTelegramBot, ApplicationSourceManual:
		return true
	}
	return false
}

func (s ApplicationStatus) Valid() bool {
	switch s {
	case ApplicationStatusQueue, ApplicationStatusInProgress, ApplicationStatusDone:
		return true
	}
	return false
}

type Application struct {
	ID               int64             `db:"id"`
	Type             ApplicationType   `db:"type"`
	Source           ApplicationSource `db:"source"`
	Status           ApplicationStatus `db:"status"`
	CustomerName     string            `db:"customer_name"`
	ContactInfo      string            `db:"contact_info"`
	ProjectName      *string           `db:"project_name"`
	BoxID            *int64            `db:"box_id"`
	SpecialProjectID *int64            `db:"special_project_id"`
	ManagerID        *int64            `db:"manager_id"`
	CreatedAt        time.Time         `db:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at"`
}

type ApplicationCreateRequest struct {
	Type             ApplicationType   `json:"type"`
	Source           ApplicationSource `json:"source"`
	CustomerName     string            `json:"customer_name"`
	ContactInfo      string            `json:"contact_info"`
	ProjectName      *string           `json:"project_name,omitempty"`
	BoxID            *int64            `json:"box_id,omitempty"`
	SpecialProjectID *int64            `json:"special_project_id,omitempty"`
}

type ApplicationFilter struct {
	Type      *ApplicationType
	Status    *ApplicationStatus
	ManagerID *int64
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}

type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ApplicationListResponse struct {
	Items      []Application `json:"items"`
	Pagination Pagination    `json:"pagination"`
}
