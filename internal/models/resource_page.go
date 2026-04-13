package models

import (
	"errors"
	"time"
)

var ErrResourcePageNotFound = errors.New("resource page not found")

type ResourcePageLink struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ResourcePage struct {
	Slug      string
	Title     string
	Content   string
	Links     []ResourcePageLink
	CreatedAt time.Time
	UpdatedAt time.Time
}
