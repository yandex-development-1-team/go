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
