package models

import (
	"time"
)

type ServiceStatus string

const (
	StatusActive   ServiceStatus = "active"
	StatusInactive ServiceStatus = "inactive"
)

// Service — услуга (БД и сервисный слой).
type Service struct {
	ID                int64
	Name              string
	Slug              string
	Description       string
	Rules             string
	Location          string
	Price             int
	Image             *string
	Status            string
	Organizer         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	BoxAvailableSlots []BoxAvailableSlot
}

type BoxCreate struct {
	Name        *string
	Slug        *string
	Description *string
	Rules       *string
	Location    *string
	Price       *int
	Image       *string
	Status      *string
	Organizer   *string
	Slots       []BoxAvailableSlot
}

type BoxUpdate struct {
	ID          int64
	Name        *string
	Description *string
	Rules       *string
	Slots       []BoxAvailableSlot
	Location    *string
	Price       *int
	Image       *string
	Status      *string
	Organizer   *string
}

// AvailableSlot — слоты по дате (дата + список времени).
type BoxAvailableSlot struct {
	Date      string
	StartTime string
	EndTime   string
}

// BoxSolutionsButton — кнопка меню «Коробочные решения».
type BoxSolutionsButton struct {
	Name  string
	Alias string
}

// BoxNewSlots - структура для INSERT всех слотов в 1 запрос
type BoxNewSlots struct {
	Date      []time.Time
	StartTime []time.Time
	EndTime   []time.Time
}

type BoxUpdateStatusResult struct {
	ID        int64
	Status    string
	UpdatedAt time.Time
}

type BoxList struct {
	Status *string
	Search *string
	Limit  int
	Offset int
	Sort   string
	Order  string
}

// BoxListResult результат списка коробок (из репозитория)
type BoxListResult struct {
	Items  []Service
	Total  int
	Limit  int
	Offset int
}
