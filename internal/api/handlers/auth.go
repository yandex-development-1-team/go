package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"

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
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        authResult.Token,
		RefreshToken: authResult.RefreshToken,
		User:         toUserResponse(authResult.User),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {

	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	token, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.RefreshResponse{Token: token})
}

func (h *AuthHandler) Logout(c *gin.Context) {

	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	c.JSON(http.StatusOK, dto.LogoutResponse{Message: "Logged out successfully"})
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
