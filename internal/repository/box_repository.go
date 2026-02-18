package repository

import (
	"context"
	"github.com/yandex-development-1-team/go/internal/handlers"
)

type Repository struct {
}

type UserRepository struct {
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

func (r Repository) GetServiceByID(ctx context.Context, serviceID int) (*handlers.Service, error) {
	return &handlers.Service{}, nil
}

func (r Repository) CreateUser(ctx context.Context, serviceID int) error {
	return nil
}
