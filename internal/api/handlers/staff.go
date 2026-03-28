package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	svcapi "github.com/yandex-development-1-team/go/internal/service/api"
)

type UsersHandler struct {
	svc *svcapi.UsersAdminService
}

func NewUsersHandler(svc *svcapi.UsersAdminService) *UsersHandler {
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
