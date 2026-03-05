package models

type Service struct {
	ID             int64           `db:"id"`          // уникальный идентификатор услуги
	Name           string          `db:"name"`        // название услуги
	Description    string          `db:"description"` // описание
	Rules          string          `db:"rules"`       // правила
	Schedule       string          `db:"schedule"`    // время проведения
	AvailableSlots []AvailableSlot // время проведения
	Type           string          `db:"type"`         // тип услуги (музей, спорт и т.д.)
	BoxSolution    bool            `db:"box_solution"` // является ли услуга коробочным решением
}

type AvailableSlot struct {
	Date      string   `db:"slot_date"`
	TimeSlots []string `db:"time_slots"`
}
