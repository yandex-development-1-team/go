package handlers

import "time"

type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginSuccessful struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

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

var rolePermissions = map[string][]string{
	"admin":   {"users:read", "users:write", "users:delete", "events:read", "events:write"},
	"manager": {"users:read", "events:read", "events:write"},
}

func PermissionsByRole(role string) []string {
	return rolePermissions[role]
}
