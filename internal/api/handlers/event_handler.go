package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
		// Помогает увидеть в логах теста, что именно не так в JSON
		fmt.Printf("[DEBUG] Bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "details": err.Error()})
		return
	}

	eventDomain, err := h.toDomain(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "message": err.Error()})
		return
	}

	created, err := h.repo.Create(c.Request.Context(), eventDomain)
	if err != nil {
		fmt.Printf("[DEBUG] Repo error: %v\n", err)
		// Если упало из-за FK (box_id), возвращаем 400
		if errors.Is(err, models.ErrBoxNotFound) || err.Error() == "box not found" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "box_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ListEvents: GET /api/v1/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	boxIDStr := c.Query("box_id")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := models.EventFilter{
		Limit:  limit,
		Offset: offset,
	}

	if boxIDStr != "" {
		id, _ := strconv.ParseInt(boxIDStr, 10, 64)
		filter.BoxID = &id
	}
	if dateFromStr != "" {
		if t, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			filter.DateFrom = &t
		}
	}
	if dateToStr != "" {
		if t, err := time.Parse("2006-01-02", dateToStr); err == nil {
			filter.DateTo = &t
		}
	}
	if status != "" {
		filter.Status = &status
	}

	items, total, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	c.JSON(http.StatusOK, dto.EventsListResponse{
		Items: items,
		Pagination: models.Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

// GetEventByID: GET /api/v1/events/:id
func (h *EventHandler) GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}

	event, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, models.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
		return
	}

	c.JSON(http.StatusOK, event)
}
