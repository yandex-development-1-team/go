package dto

import "github.com/yandex-development-1-team/go/internal/service/models" // Замените на ваш путь

// ResourcePageUpdateRequest DTO for updating a resource page.
// Using pointers to allow optional fields.
type ResourcePageUpdateRequest struct {
	Title   *string        `json:"title,omitempty"`
	Content *string        `json:"content,omitempty"`
	Links   *[]models.Link `json:"links,omitempty"` // Optional array of links
}

// ResourcePagePublic DTO for the public endpoint.
// Excludes slug and updated_at.
type ResourcePagePublic struct {
	Title   string        `json:"title"`
	Content string        `json:"content"`
	Links   []models.Link `json:"links"`
}
