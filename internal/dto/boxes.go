package dto

// BoxListItem — элемент списка GET /api/v1/boxes
type BoxListItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	BoxSolution bool   `json:"box_solution"`
}
