package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

type AuthorisationService interface {
	Login(context.Context, string, string) (*models.AuthResult, error)
}

type AuthHandler struct {
	service AuthorisationService
}

func NewAuthHandler(service AuthorisationService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (ah *AuthHandler) HandleLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	authResult, err := ah.service.Login(c.Request.Context(), req.Login, req.Password)
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
