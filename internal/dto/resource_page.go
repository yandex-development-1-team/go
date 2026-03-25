package dto

import "github.com/yandex-development-1-team/go/internal/models"

type ResourcePageUpdateRequest struct {
	Slug    string                     `json:"slug,omitempty"`
	Title   *string                    `json:"title,omitempty"`
	Content *string                    `json:"content,omitempty"`
	Links   *[]models.ResourcePageLink `json:"links,omitempty"`
}

type ResourcePageResponse struct {
	Title   string                     `json:"title,omitempty"`
	Content *string                    `json:"content,omitempty"`
	Links   *[]models.ResourcePageLink `json:"links,omitempty"`
}

type ResourcePageResponsePublic struct {
	Title   string                     `json:"title,omitempty"`
	Content *string                    `json:"content,omitempty"`
	Links   *[]models.ResourcePageLink `json:"links,omitempty"`
}

type ResourcePagePublic struct {
	Title   string                    `json:"title"`
	Content string                    `json:"content"`
	Links   []models.ResourcePageLink `json:"links"`
}
