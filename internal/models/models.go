package models

type GetBoxSolutionsResponse struct {
	Items []BoxSolution
}

type BoxSolution struct {
	ID             int64
	Name           string
	Description    string
	AvailableSlots []AvailableSlot
}

type AvailableSlot struct {
	Date      string
	TimeSlots []string
}

type BoxSolutionButtons struct {
	Description string
	Buttons     []Button
}

type Button struct {
	Alias string
	Name  string
}

type GetDetailsForBoxSolutionRequest struct {
	UserID string
	Button Button
}
