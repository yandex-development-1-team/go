package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type AuthHandler struct {
	svc *svcapi.AuthService
}

func NewAuthHandler(s *svcapi.AuthService) *AuthHandler {
	return &AuthHandler{svc: s}
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	Token string `json:"token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutResponse struct {
	Message string `json:"message"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeUnauthorized(w, "invalid_refresh+token", "invalid refresh token")
		return
	}

	token, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		writeUnauthorized(w, "invalid_refresh+token", "invalid or expired refresh token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(refreshResponse{Token: token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeUnauthorized(w, "invalid_refresh+token", "invalid refresh token")
		return
	}

	_ = h.svc.Logout(ctx, req.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(logoutResponse{Message: "Logged out successfully"})
}

func writeUnauthorized(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	resp := map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": []map[string]string{},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}
