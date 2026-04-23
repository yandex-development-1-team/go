package dto

import "time"

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

type DashboardOverview struct {
	NewApplications        int64 `json:"new_applications"         db:"new_applications"`
	InProgressApplications int64 `json:"in_progress_applications" db:"in_progress_applications"`
}

type DashboardManagerStats struct {
	InProgress int64 `json:"in_progress" db:"in_progress"`
	Processed  int64 `json:"processed"   db:"processed"`
}

type ApplicationShort struct {
	TelegramNick string `json:"telegram_nick" db:"tg_account"`
	CustomerName string `json:"name"          db:"customer_name"`
	ServiceType  string `json:"service_type"  db:"service_type"`
	ServiceName  string `json:"service_name"  db:"service_name"`
	Status       string `json:"status"        db:"status"`
	CreatedAt    string `json:"created_at"    db:"created_at"`
}

type DashboardResponse struct {
	Overview     DashboardOverview     `json:"overview"`
	ManagerStats DashboardManagerStats `json:"manager_stats"`
	Applications []ApplicationShort    `json:"applications"`
}
