package models

// Service — услуга (БД и сервисный слой).
type Service struct {
	ID             int64           `db:"id"`
	Name           string          `db:"name"`
	Description    string          `db:"description"`
	Rules          string          `db:"rules"`
	Schedule       string          `db:"schedule"`
	Type           string          `db:"type"`
	BoxSolution    bool            `db:"box_solution"`
	AvailableSlots []AvailableSlot `db:"-"`
}

// AvailableSlot — слоты по дате (дата + список времени).
type AvailableSlot struct {
	Date      string   `db:"slot_date"`
	TimeSlots []string `db:"time_slots"`
}

// BoxSolutionsButton — кнопка меню «Коробочные решения».
type BoxSolutionsButton struct {
	Name  string
	Alias string
}
