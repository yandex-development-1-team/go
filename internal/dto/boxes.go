package dto

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
