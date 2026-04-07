package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/yandex-development-1-team/go/internal/models"
)

// Errors
var (
	ErrServiceNotFound = errors.New("service not found")
	ErrIncorrectData   = errors.New("handler with incorrect data")
	ErrInvalidField    = errors.New("handler with an invalid field")
)

// ServiceRepo defines the data access layer interface for service operations
type ServiceRepo interface {
	GetServiceByID(ctx context.Context, serviceID int64) (*models.Service, error)
}

// DetailService provides logic for service detail
type DetailService struct {
	repo ServiceRepo
}

// NewDetailService creates a new instance of the 'DetailService'
func NewDetailService(repo ServiceRepo) *DetailService {
	return &DetailService{repo: repo}
}

// GetByID retrieves a service from the database by its ID
func (s *DetailService) GetByID(ctx context.Context, serviceID int64) (*models.Service, error) {
	service, err := s.repo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return service, nil
}

// ParseServiceID returns the service ID
func (s *DetailService) ParseServiceID(callbackData string) (int64, error) {
	parts := strings.Split(callbackData, ":")

	if len(parts) != 3 {
		return 0, ErrIncorrectData
	}

	if parts[0] != "info" {
		return 0, ErrInvalidField
	}

	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid service id format: %w", err)
	}
	return id, nil
}

// GetDisplayName returns the service name for display
func (s *DetailService) GetDisplayName(service *models.Service) string {
	if service.Name == "" {
		return "Прочее"
	}
	return service.Name
}
