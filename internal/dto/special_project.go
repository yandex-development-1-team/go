package dto

import "time"

type SpecialProjectCreateRequest struct {
	Title         string  `json:"title"`
	Description   *string `json:"description,omitempty"`
	Image         string  `json:"image,omitempty"`
	IsActiveInBot bool    `json:"is_active_in_bot"`
}

// SpecialProjectListItem for GET /special-projects
type SpecialProjectListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	IsActiveInBot bool   `json:"is_active_in_bot"`
}

// SpecialProjectResponse for GET /special-projects/{id}
type SpecialProjectResponse struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	Image         string    `json:"image,omitempty"`
	IsActiveInBot bool      `json:"is_active_in_bot"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
