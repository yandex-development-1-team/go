package dto

import "time"

type UserCreateRequest struct {
	TelegramNick *string  `json:"telegram_nick,omitempty"`
	Name         string   `json:"name" binding:"required,min=1"`
	Email        string   `json:"email" binding:"required,email"`
	Role         string   `json:"role" binding:"required"`
	Status       string   `json:"status,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
}

type UserUpdateRequest struct {
	Name         *string   `json:"name,omitempty"`
	Email        *string   `json:"email,omitempty"`
	Role         *string   `json:"role,omitempty"`
	Status       *string   `json:"status,omitempty"`
	Permissions  *[]string `json:"permissions,omitempty"`
	TelegramNick *string   `json:"telegram_nick,omitempty"`
}

type BlockResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
