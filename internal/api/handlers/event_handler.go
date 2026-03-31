package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository/postgres"
)

type EventHandler struct {
	repo repository.EventRepository
}

func NewEventHandler(repo repository.EventRepository) *EventHandler {
	return &EventHandler{repo: repo}
}

func (h *EventHandler) toDomain(req *dto.EventCreateRequest) (*models.Event, error) {
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return &models.Event{
		BoxID:      req.BoxID,
		Date:       parsedDate,
		Time:       req.Time,
		TotalSlots: req.TotalSlots,
		Status:     models.EventStatusActive,
	}, nil
}

// CreateEvent: POST /api/v1/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req dto.EventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.WriteErrorGin(c, models.ErrInvalidInput)
		return
	}

	eventDomain, err := h.toDomain(&req)
	if err != nil {
		apierrors.WriteErrorGin(c, models.ErrInvalidInput)
		return
	}

	created, err := h.repo.Create(c.Request.Context(), eventDomain)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListEvents: GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	var query dto.EventListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierrors.WriteErrorGin(c, models.ErrInvalidInput)
		return
	}

	filter := models.EventFilter{
		BoxID:  query.BoxID,
		Status: query.Status,
		Limit:  query.Limit,  // Теперь берется из DTO с дефолтом
		Offset: query.Offset, // Теперь берется из DTO с дефолтом
	}

	// Парсинг дат оставляем простым
	if query.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", query.DateFrom); err == nil {
			filter.DateFrom = &t
		}
	}
	if query.DateTo != "" {
		if t, err := time.Parse("2006-01-02", query.DateTo); err == nil {
			filter.DateTo = &t
		}
	}

	items, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.EventsListResponse{
		Items: items,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  filter.Limit,
			Offset: filter.Offset,
		},
	})
}

// GetEventByID: GET /api/v1/events/:id
func (h *EventHandler) GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apierrors.WriteErrorGin(c, models.ErrInvalidInput)
		return
	}

	event, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.JSON(http.StatusOK, event)
}
