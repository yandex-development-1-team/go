package dto

import "time"

type UserCreateRequest struct {
	FirstName   string  `json:"first_name"          binding:"required,min=1,max=255"`
	LastName    string  `json:"last_name"            binding:"required,min=1,max=255"`
	Email       string  `json:"email"                binding:"required,email,max=255"`
	Role        string  `json:"role"                 binding:"required,oneof=admin manager_1 manager_2 manager_3 user"`
	Status      string  `json:"status,omitempty"     binding:"omitempty,oneof=active blocked invited"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,max=50"`
	Image       *string `json:"image,omitempty"      binding:"omitempty,httpurl,max=500"`
	Department  *string `json:"department,omitempty" binding:"omitempty,max=255"`
	Position    *string `json:"position,omitempty"   binding:"omitempty,max=255"`
	Supervisor  *string `json:"supervisor,omitempty" binding:"omitempty,max=255"`
	Address     *string `json:"address,omitempty"    binding:"omitempty,max=500"`
	SecondName  *string `json:"second_name,omitempty" binding:"omitempty,max=255"`
}

type UserUpdateRequest struct {
	FirstName   *string `json:"first_name,omitempty"  binding:"omitempty,min=1,max=255"`
	LastName    *string `json:"last_name,omitempty"   binding:"omitempty,min=1,max=255"`
	Email       *string `json:"email,omitempty"       binding:"omitempty,email,max=255"`
	Role        *string `json:"role,omitempty"        binding:"omitempty,oneof=admin manager_1 manager_2 manager_3 user"`
	Status      *string `json:"status,omitempty"      binding:"omitempty,oneof=active blocked invited"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,max=50"`
	Image       *string `json:"image,omitempty"       binding:"omitempty,httpurl,max=500"`
	Department  *string `json:"department,omitempty"  binding:"omitempty,max=255"`
	Position    *string `json:"position,omitempty"    binding:"omitempty,max=255"`
	Supervisor  *string `json:"supervisor,omitempty"  binding:"omitempty,max=255"`
	Address     *string `json:"address,omitempty"     binding:"omitempty,max=500"`
	SecondName  *string `json:"second_name,omitempty" binding:"omitempty,max=255"`
}

type UpdateStatusResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active blocked"`
}

type UserListItem struct {
	ID           int64     `json:"id"`
	TelegramNick string    `json:"telegram_nick"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	SecondName   string    `json:"second_name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Department   string    `json:"department"`
	Supervisor   string    `json:"supervisor"`
	Position     string    `json:"position"`
	PhoneNumber  string    `json:"phone_number"`
	Email        string    `json:"email"`
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
	FirstName     string             `json:"first_name"`
	LastName      string             `json:"last_name"`
	SecondName    string             `json:"second_name"`
	Email         string             `json:"email"`
	PhoneNumber   string             `json:"phone_number"`
	Role          string             `json:"role"`
	Status        string             `json:"status"`
	Department    string             `json:"department"`
	Position      string             `json:"position"`
	Supervisor    string             `json:"supervisor"`
	Address       string             `json:"address"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Bookings      []UserBookingItem  `json:"bookings"`
	VisitHistory  []VisitHistoryItem `json:"visit_history"`
	FavoriteBoxes []int64            `json:"favorite_boxes"`
	Image         string             `json:"image"`
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
