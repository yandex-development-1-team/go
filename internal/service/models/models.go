package models

import "time"

type Service struct {
	ID             int64
	Name           string
	Description    string
	AvailableSlots []AvailableSlot
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}

type BoxSolutionsButton struct {
	Name  string
	Alias string
}

// --- Domain Model ---
type SpecialProject struct {
	ID            int64
	Title         string
	Description   *string
	Image         string
	IsActiveInBot bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Link represents a single link in the links array.
type Link struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ResourcePage represents a full resource page entity.
type ResourcePage struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Links     []Link `json:"links"`
	UpdatedAt string `json:"updated_at"`
}

// ResourcePageSummary represents a summary for listing pages.
type ResourcePageSummary struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
}
