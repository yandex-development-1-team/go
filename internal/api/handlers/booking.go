package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	service "github.com/yandex-development-1-team/go/internal/service/api"
)

type BookingHandler struct {
	svc *service.BookingsService
}

func NewBookingHandler(svc *service.BookingsService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

func (h *BookingHandler) BookingsById(c *gin.Context) {
	var id dto.BookingsID
	if err := c.ShouldBindUri(&id); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	booking, err := h.svc.GetBookingById(c.Request.Context(), id.ID)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toBookingDetailResponse(booking))
}

func (h *BookingHandler) BookingsList(c *gin.Context) {
	var query dto.ApplicationListRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	list, err := h.svc.GetBookingsList(c.Request.Context(), toAppFilter(&query))
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toBookingsListResponse(list))
}

func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	var id dto.BookingsID
	if err := c.ShouldBindUri(&id); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	var status dto.ApplicationUpdateStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Некорректные данные"})
		return
	}

	app, err := h.svc.UpdateBookingStatus(c.Request.Context(), id.ID, status.Status)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, toBookingDetailResponse(app))
}

func (h *BookingHandler) DeleteBooking(c *gin.Context) {
	var uri models.ApplicationURI
	if err := c.ShouldBindUri(&uri); err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest, []string{"Неверные параметры запроса"})
		return
	}

	err := h.svc.DeleteBooking(c.Request.Context(), uri.ID)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
func toBookingsListResponse(result *models.BookingList) dto.BookingListResponse {
	items := make([]dto.BookingListItem, len(result.Items))
	for i, app := range result.Items {
		items[i] = dto.BookingListItem{
			ID:           app.ID,
			Status:       app.Status,
			ServiceName:  app.ServiceName,
			ManagerID:    app.ManagerID,
			ManagerName:  app.ManagerName,
			CustomerName: app.GuestName,
			ContactInfo:  app.GuestContact,
			CreatedAt:    app.CreatedAt,
		}
	}

	return dto.BookingListResponse{
		Items: items,
		Pagination: dto.Pagination{
			Total:  result.Total,
			Limit:  result.Limit,
			Offset: result.Offset,
		},
	}
}

func toBookingDetailResponse(b *models.BookingAPI) dto.BookingDetailResponse {
	return dto.BookingDetailResponse{
		ID:                b.ID,
		UserID:            b.UserID,
		ServiceID:         b.ServiceID,
		ServiceName:       b.ServiceName,
		BookingDate:       b.BookingDate,
		BookingTime:       b.BookingTime,
		GuestName:         b.GuestName,
		GuestOrganization: b.GuestOrganization,
		GuestContact:      b.GuestContact,
		GuestPosition:     b.GuestPosition,
		Status:            b.Status,
		ManagerID:         b.ManagerID,
		ManagerName:       b.ManagerName,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}
