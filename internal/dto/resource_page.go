package dto

import (
	"encoding/json"
	"time"
)

type ResourcePageLink struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url" binding:"required,url"`
}

type ResourcePageUpdateRequest struct {
	Title   string             `json:"title"`
	Content string             `json:"content"`
	Links   []ResourcePageLink `json:"links" binding:"dive"`
}

type ResourcePageResponse struct {
	Slug      string             `json:"slug"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Links     []ResourcePageLink `json:"links"`
	UpdatedAt string             `json:"updated_at"`
}

type ResourcePageDB struct {
	Slug      string          `db:"slug"`
	Title     string          `db:"title"`
	Content   string          `db:"content"`
	Links     json.RawMessage `db:"links"`
	UpdatedAt time.Time       `db:"updated_at"`
}

type ResourcePagePublicResponse struct {
	Title   string             `json:"title"`
	Content string             `json:"content"`
	Links   []ResourcePageLink `json:"links"`
}
