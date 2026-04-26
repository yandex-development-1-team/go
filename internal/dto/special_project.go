package dto

import (
	"time"
)

type SpecialProjectCreateRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Description string  `json:"description" binding:"required,max=1000"`
	Image       *string `json:"image,omitempty" binding:"omitempty,httpurl,max=500"`
	Status      string  `json:"status" binding:"required,oneof=active inactive"`
}

type SpecialProjectUpdateRequest struct {
	Title       *string `json:"title,omitempty"       binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
	Image       *string `json:"image,omitempty"       binding:"omitempty,httpurl,max=500"`
	Status      *string `json:"status,omitempty"      binding:"omitempty,oneof=active inactive"`
}

type SpecialProjectListItem struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       *string   `json:"image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SpecialProjectResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
