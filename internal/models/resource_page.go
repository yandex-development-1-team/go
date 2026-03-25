package models

import (
	"encoding/json"
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
	Slug      string             `json:"slug"`
	Title     string             `json:"title"`
	Content   string             `json:"content"`
	Links     []ResourcePageLink `json:"-"`
	LinksJSON json.RawMessage    `json:"links"`
	UpdatedAt string             `json:"updated_at"`
}

type ResourcePageDB struct {
	Slug      string          `db:"slug"`
	Title     string          `db:"title"`
	Content   *string         `db:"content"`
	LinksJSON json.RawMessage `db:"links"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}
