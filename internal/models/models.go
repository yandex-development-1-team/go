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
	ManagerID    int64
	InviteToken  string
	Permissions  []string
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

type BookingExt struct {
	ID                int64      `db:"id"`
	UserID            int64      `db:"user_id"`
	ServiceID         int16      `db:"service_id"`
	ServiceName       string     `db:"service_name"`
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

type UserSession struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	CurrentState string                 `json:"current_state"`
	StateData    map[string]interface{} `json:"state_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type (
	ApplicationType   string
	ApplicationSource string
	ApplicationStatus string
)

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
	ID               int64             `db:"id" json:"id"`
	Type             ApplicationType   `db:"type" json:"type"`
	Source           ApplicationSource `db:"source" json:"source"`
	Status           ApplicationStatus `db:"status" json:"status"`
	CustomerName     string            `db:"customer_name" json:"customer_name"`
	ContactInfo      string            `db:"contact_info" json:"contact_info"`
	ProjectName      *string           `db:"project_name" json:"project_name,omitempty"`
	BoxID            *int64            `db:"box_id" json:"box_id,omitempty"`
	SpecialProjectID *int64            `db:"special_project_id" json:"special_project_id,omitempty"`
	ManagerID        *int64            `db:"manager_id" json:"manager_id,omitempty"`
	ManagerName      *string           `db:"manager_name" json:"manager_name,omitempty"`
	CreatedAt        time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time         `db:"updated_at" json:"updated_at"`
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
	Type         *ApplicationType
	Status       *ApplicationStatus
	ManagerID    *int64
	CustomerName string
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
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

type ApplicationUpdateRequest struct {
	Status           *ApplicationStatus `json:"status,omitempty"`
	ContactInfo      *string            `json:"contact_info,omitempty"`
	BoxID            *int64             `json:"box_id,omitempty"`
	SpecialProjectID *int64             `json:"special_project_id,omitempty"`
}

func (r *ApplicationUpdateRequest) HasUpdates() bool {
	return r.Status != nil || r.ContactInfo != nil || r.BoxID != nil || r.SpecialProjectID != nil
}

type RefreshResponse struct {
	Token        string
	RefreshToken string
}
