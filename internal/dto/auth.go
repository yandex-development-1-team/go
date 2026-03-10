package dto

import "time"

// LoginRequest — тело запроса POST /api/v1/auth/login.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginResponse — успешный ответ логина (token + user).
type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// UserResponse — пользователь в ответах API (auth, users).
type UserResponse struct {
	ID           int64     `json:"id"`
	TelegramNick string    `json:"telegram_nick"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	Permissions  []string  `json:"permissions"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefreshRequest — body для POST /api/v1/auth/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse — ответ с новым access token
type RefreshResponse struct {
	Token string `json:"token"`
}

// LogoutRequest — body для POST /api/v1/auth/logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutResponse — ответ после logout
type LogoutResponse struct {
	Message string `json:"message"`
}
