package dto

import "time"

type BoxListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	BoxSolution bool   `json:"box_solution"`
}

type BoxDetail struct {
	ID             int64              `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description,omitempty"`
	Rules          string             `json:"rules,omitempty"`
	Schedule       string             `json:"schedule,omitempty"`
	Type           string             `json:"type,omitempty"`
	BoxSolution    bool               `json:"box_solution"`
	AvailableSlots []BoxAvailableSlot `json:"available_slots,omitempty"`
}

type BoxAvailableSlot struct {
	Date      string   `json:"date"`
	TimeSlots []string `json:"time_slots"`
}

type BoxUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	BoxSolution *bool   `json:"box_solution,omitempty"`
}

type BoxStatusResponse struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
