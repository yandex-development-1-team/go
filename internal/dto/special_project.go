package dto

import (
	"time"

	"github.com/yandex-development-1-team/go/internal/models"
)

type SpecialProjectCreateRequest struct {
	Title       string               `json:"title" binding:"required,min=1,max=255"`
	Description *string              `json:"description,omitempty" binding:"omitempty,max=1000"`
	Image       string               `json:"image,omitempty" binding:"omitempty,httpurl,max=500"`
	Status      models.ServiceStatus `json:"status" binding:"required,oneof=active inactive"`
}

type SpecialProjectListItem struct {
	ID          int64                `json:"id,omitempty"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	Image       string               `json:"image,omitempty"`
	Status      models.ServiceStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at,omitempty"`
	UpdatedAt   time.Time            `json:"updated_at,omitempty"`
}

type SpecialProjectResponse struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Description *string              `json:"description,omitempty"`
	Image       string               `json:"image,omitempty"`
	Status      models.ServiceStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type SpecialProjectListResponse struct {
	Items      []SpecialProjectListItem `json:"items"`
	Pagination Pagination               `json:"pagination"`
}

type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
