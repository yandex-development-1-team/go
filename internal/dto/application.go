package dto

import (
	"time"
)

const (
	DefaultApplicationLimit = 20
	MaxApplicationLimit     = 100
)

type ApplicationUpdateStatus struct {
	Status string `json:"status" binding:"required,oneof=pending confirmed cancelled"`
}

type Application struct {
	ID           int64     `json:"id"`
	Status       string    `json:"status"`
	ManagerID    int64     `json:"manager_id"`
	ManagerName  string    `json:"manager_name"`
	FormAnswerId string    `json:"form_answer_id"`
	CustomerName string    `json:"customer_name"`
	ContactInfo  string    `json:"contact_info"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ApplicationDB struct {
	ID           int64     `db:"id"`
	Status       string    `db:"status"`
	ManagerID    int64     `db:"manager_id"`
	ManagerName  string    `db:"manager_name"`
	FormAnswerId string    `db:"form_answer_id"`
	CustomerName string    `db:"customer_name"`
	ContactInfo  string    `db:"contact_info"`
	Description  string    `db:"description"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type CreateApplicationRequest struct {
	Answer struct {
		Data struct {
			FirstName struct {
				Value string `json:"value"`
			} `json:"first_name"`
			LastName struct {
				Value string `json:"value"`
			} `json:"last_name"`
			Telegram struct {
				Value string `json:"value"`
			} `json:"telegram"`
			Description struct {
				Value string `json:"value"`
			} `json:"description"`
		} `json:"data"`
	} `json:"answer"`
}

type ApplicationListRequest struct {
	Status       *string `form:"status" binding:"omitempty,oneof=pending confirmed cancelled"`
	ManagerID    *int64  `form:"manager_id" binding:"omitempty,min=1"`
	CustomerName *string `form:"customer_name"`
	Limit        int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset       int     `form:"offset" binding:"omitempty,min=0"`
}

type ApplicationListItem struct {
	ID           int64     `json:"id"`
	Status       string    `json:"status"`
	ManagerID    int64     `json:"manager_id"`
	ManagerName  string    `json:"manager_name"`
	CustomerName string    `json:"customer_name"`
	ContactInfo  string    `json:"contact_info"`
	CreatedAt    time.Time `json:"created_at"`
}

type ApplicationListResponse struct {
	Items      []ApplicationListItem `json:"items"`
	Pagination Pagination            `json:"pagination"`
}

type ApplicationRow struct {
	ID           int64     `db:"id"`
	Status       string    `db:"status"`
	ManagerID    int64     `db:"manager_id"`
	ManagerName  string    `db:"manager_name"`
	CustomerName string    `db:"customer_name"`
	ContactInfo  string    `db:"contact_info"`
	CreatedAt    time.Time `db:"created_at"`
	Total        int       `db:"total"`
}
