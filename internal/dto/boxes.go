package dto

import "time"

// BoxListItem — элемент списка GET /api/v1/boxes
type BoxListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	BoxSolution bool   `json:"box_solution"`
}

// BoxDetail — полная информация GET /api/v1/boxes/:id
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

// BoxAvailableSlot — слот доступности для коробки.
type BoxAvailableSlot struct {
	Date      string   `json:"date"`
	TimeSlots []string `json:"time_slots"`
}

// BoxUpdateRequest — тело запроса PUT /api/v1/boxes/:id (все поля optional)
type BoxUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	BoxSolution *bool   `json:"box_solution,omitempty"`
}

// BoxStatusResponse — ответ PUT /api/v1/boxes/:id/status
type BoxStatusResponse struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
