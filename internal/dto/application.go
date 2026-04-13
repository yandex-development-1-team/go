package dto

import "github.com/yandex-development-1-team/go/internal/models"

const (
	DefaultApplicationLimit = 20
	MaxApplicationLimit     = 100
)

type ApplicationListQuery struct {
	Status       *string `form:"status" binding:"omitempty,oneof=queue in_progress done"`
	Type         *string `form:"type" binding:"omitempty,oneof=box special_project"`
	ManagerID    *int64  `form:"manager_id" binding:"omitempty,min=1"`
	CustomerName string  `form:"customer_name"`
	Limit        int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset       int     `form:"offset" binding:"omitempty,min=0"`
}

type ApplicationListResponse struct {
	Items      []models.Application `json:"items"`
	Pagination Pagination           `json:"pagination"`
}
