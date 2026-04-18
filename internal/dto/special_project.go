package dto

import (
	"time"

	"github.com/yandex-development-1-team/go/internal/models"
)

type SpecialProjectCreateRequest struct {
	Title       string               `json:"title" binding:"required,min=1"`
	Description *string              `json:"description,omitempty"`
	Image       string               `json:"image,omitempty"`
	Status      models.ServiceStatus `json:"status"`
}

type SpecialProjectListItem struct {
	ID     int64                `json:"id"`
	Title  string               `json:"title"`
	Status models.ServiceStatus `json:"status"`
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
