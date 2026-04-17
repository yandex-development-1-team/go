package dto

import "time"

type UserCreateRequest struct {
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email" binding:"required,email"`
	Role       string  `json:"role" binding:"required"`
	Status     string  `json:"status,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Department *string `json:"department,omitempty"`
	Position   *string `json:"position,omitempty"`
	SecondName *string `json:"second_name,omitempty"`
}

type UserUpdateRequest struct {
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	Email      *string `json:"email,omitempty"`
	Role       *string `json:"role,omitempty"`
	Status     *string `json:"status,omitempty"`
	Phone      *string `json:"phone_number,omitempty"`
	Department *string `json:"department,omitempty"`
	Position   *string `json:"position,omitempty"`
	SecondName *string `json:"second_name,omitempty"`
}

type BlockResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListItem struct {
	ID           int64     `json:"id"`
	TelegramNick string    `json:"telegram_nick"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserListResponse struct {
	Items      []UserListItem `json:"items"`
	Pagination Pagination     `json:"pagination"`
}

type UserBookingItem struct {
	ID          int64     `json:"id"`
	ServiceName string    `json:"service_name"`
	BookingDate time.Time `json:"booking_date"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type VisitHistoryItem struct {
	BoxName   string    `json:"box_name"`
	VisitedAt time.Time `json:"visited_at"`
}

type UserWithDetails struct {
	ID            int64              `json:"id"`
	TelegramNick  string             `json:"telegram_nick"`
	Name          string             `json:"name"`
	LastName      string             `json:"last_name"`
	SecondName    string             `json:"second_name"`
	Email         string             `json:"email"`
	PhoneNumber   string             `json:"phone_number"`
	Role          string             `json:"role"`
	Status        string             `json:"status"`
	Department    string             `json:"department"`
	Position      string             `json:"position"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Bookings      []UserBookingItem  `json:"bookings"`
	VisitHistory  []VisitHistoryItem `json:"visit_history"`
	FavoriteBoxes []int64            `json:"favorite_boxes"`
}
