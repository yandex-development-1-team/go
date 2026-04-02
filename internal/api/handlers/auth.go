package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

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
		log.Printf("login: %v", err)
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        authResult.Token,
		RefreshToken: authResult.RefreshToken,
		User:         toUserResponse(authResult.User),
	})
}

func (h *AuthHandler) HandleRefresh(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	token, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusUnauthorized, []string{"Некорректный или просроченный refresh token"})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshResponse{Token: token})
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	authResult, err := h.svc.Register(c.Request.Context(), &models.UserAPI{
		Name:        req.Name,
		LastName:    req.LastName,
		Email:       req.Email,
		InviteToken: req.InviteToken,
	}, req.Password)
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

func (h *AuthHandler) HandleLogout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	_ = h.svc.Logout(ctx, req.RefreshToken)

	c.JSON(http.StatusOK, dto.LogoutResponse{Message: "Logged out successfully"})
}

func (h *AuthHandler) HandleForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	err := h.svc.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		log.Printf("forgot password: %v", err)
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Если email существует, ссылка для восстановления пароля отправлена"})
}

func (h *AuthHandler) HandleResetPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Токен не указан"})
		return
	}

	newPassword := c.Query("password")
	if newPassword == "" || len(newPassword) < 8 {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Пароль должен содержать минимум 8 символов"})
		return
	}

	err := h.svc.ResetPassword(c.Request.Context(), token, newPassword)
	if err != nil {
		log.Printf("reset password: %v", err)
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменен"})
}

func toUserResponse(user *models.UserAPI) dto.UserResponse {
	if user == nil {
		return dto.UserResponse{}
	}
	return dto.UserResponse{
		ID:           user.ID,
		TelegramNick: user.TelegramNick,
		Name:         user.Name,
		LastName:     user.LastName,
		SecondName:   user.SecondName,
		Email:        user.Email,
		PhoneNumber:  user.PhoneNumber,
		Role:         user.Role,
		Status:       user.Status,
		Department:   user.Department,
		Position:     user.Position,
		ManagerID:    user.ManagerID,
		InviteToken:  user.InviteToken,
		Permissions:  user.Permissions,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
