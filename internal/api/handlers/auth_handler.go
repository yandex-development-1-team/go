package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
	var loginData LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		// logger.Error("bad_request", zap.Error(err), zap.String("operation", "HandleLogin"))
		errorWriter(c, models.ErrInvalidInput)
		return
		// if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
		// 	// logger.Info("") // дописать
		// 	BadRequestError(w, err.Error())
		// 	return
	}

	authResult, err := ah.service.Login(c.Request.Context(), loginData.Login, loginData.Password)
	if err != nil {
		// обработка и логирование ошибок должны быть в middleware
		errorWriter(c, err)
		return
	}

	userResponse := toUserResponse(authResult.User)

	response := LoginSuccessful{
		Token:        authResult.Token,
		RefreshToken: authResult.RefreshToken,
		User:         userResponse,
	}

	c.JSON(http.StatusOK, response)
}

func toUserResponse(user *models.UserAPI) UserResponse {
	userResponse := UserResponse{
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

	return userResponse
}

func errorWriter(c *gin.Context, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidCredentials),
		errors.Is(err, models.ErrUserNotFound):
		// logger.Info("wrong_credentials", zap.Error(err), zap.String("operation", "HandleLogin"))
		c.JSON(http.StatusUnauthorized, UnauthorizedError("неверный логин или пароль"))

	case errors.Is(err, models.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, BadRequestError())

	case errors.Is(err, models.ErrUserBlocked):
		// logger.Info("user_blocked", zap.Error(err), zap.String("operation", "HandleLogin"))
		c.JSON(http.StatusForbidden, ForbiddenError())

	default:
		// logger.Info("internal_error", zap.Error(err), zap.String("operation", "HandleLogin"))
		c.JSON(http.StatusInternalServerError, InternalServerError())
	}
}
