package db_models

type Box struct {
	Name           string
	Description    string
	AvailableSlots []AvailableSlot
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}
