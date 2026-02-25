package models

type Service struct {
	ID          int    // Уникальный идентификатор услуги
	Name        string // название услуги
	Description string // описание
	Rules       string // правила
	Schedule    string // время проведения
	Type        string // Тип услуги (музей, спорт и т.д.)
	BoxID       int    // ID бокса/категории для кнопки "Назад"
}
