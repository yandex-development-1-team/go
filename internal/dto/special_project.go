package dto

import "time"

type SpecialProjectCreateRequest struct {
	Title         string  `json:"title" binding:"required,min=1"`
	Description   *string `json:"description,omitempty"`
	Image         string  `json:"image,omitempty"`
	IsActiveInBot bool    `json:"is_active_in_bot"`
}

type SpecialProjectListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	IsActiveInBot bool   `json:"is_active_in_bot"`
}

type SpecialProjectResponse struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Image         string    `json:"image,omitempty"`
	IsActiveInBot bool      `json:"is_active_in_bot"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
