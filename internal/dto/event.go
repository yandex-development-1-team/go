package dto

import "github.com/yandex-development-1-team/go/internal/models"

type EventCreateRequest struct {
	BoxID      int64  `json:"box_id" binding:"required"`
	Date       string `json:"date" binding:"required"`
	Time       string `json:"time" binding:"required"`
	TotalSlots int    `json:"total_slots" binding:"required,min=1"`
}

type EventResponse struct {
	models.Event
}

type EventListItemResponse struct {
	models.EventListItem
}

type EventsListResponse struct {
	Items      []models.EventListItem `json:"items"`
	Pagination models.Pagination      `json:"pagination"`
}
