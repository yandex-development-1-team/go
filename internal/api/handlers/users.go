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
