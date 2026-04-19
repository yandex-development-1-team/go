package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

type UsersHandler struct {
	svc *apiService.UsersAdminService
}

func NewUsersHandler(svc *apiService.UsersAdminService) *UsersHandler {
	return &UsersHandler{svc: svc}
}

func (h *UsersHandler) Create(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}
	user, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *UsersHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный id"})
		return
	}
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}
	user, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UsersHandler) Block(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный id"})
		return
	}
	user, err := h.svc.Block(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.BlockResponse{
		ID:        user.ID,
		Status:    "blocked",
		UpdatedAt: user.UpdatedAt,
	})
}

type UserHandler struct {
	svc *apiService.UserService
}

func NewUserHandler(svc *apiService.UserService) *UserHandler {

	return &UserHandler{svc: svc}
}

func (h *UserHandler) List(c *gin.Context) {
	role := c.Query("role")
	status := c.Query("status")
	search := c.Query("search")

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, total, err := h.svc.List(c.Request.Context(), role, status, search, limit, offset)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.UserListResponse{
		Items: items,
		Pagination: dto.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректный идентификатор пользователя"})
		return
	}

	user, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrUserNotFound {
			apierrors.WriteErrorMessagesGin(c, http.StatusNotFound, []string{"Пользователь не найден"})
			return
		}
		apierrors.WriteErrorGin(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}
