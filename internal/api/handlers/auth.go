package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type AuthorisationService interface {
	Login(context.Context, string, string) (*models.AuthResult, error)
}

type AuthHandler struct {
	svc *svcapi.AuthService
}

func NewAuthHandler(s *svcapi.AuthService) *AuthHandler {
	return &AuthHandler{svc: s}
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	authResult, err := h.svc.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        authResult.Token,
		RefreshToken: authResult.RefreshToken,
		User:         toUserResponse(authResult.User),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeUnauthorized(w, "invalid_refresh_token", "invalid refresh token")
		return
	}

	token, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		writeUnauthorized(w, "invalid_refresh_token", "invalid or expired refresh token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.RefreshResponse{Token: token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeUnauthorized(w, "invalid_refresh_token", "invalid refresh token")
		return
	}

	_ = h.svc.Logout(ctx, req.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.LogoutResponse{Message: "Logged out successfully"})
}

func toUserResponse(user *models.UserAPI) dto.UserResponse {
	if user == nil {
		return dto.UserResponse{}
	}
	return dto.UserResponse{
		ID:           user.ID,
		TelegramNick: user.TelegramNick,
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		Permissions:  user.Permissions,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
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
