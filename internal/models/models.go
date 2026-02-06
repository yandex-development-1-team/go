package models

type GetBoxSolutionsResponse struct {
	Items []Box
}

type Box struct {
	Name           string
	Description    string
	AvailableSlots []AvailableSlot
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}
