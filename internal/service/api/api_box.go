package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

// BoxLister returns box solutions (services with box_solution=true) from storage.
type BoxLister interface {
	GetServices(ctx context.Context, telegramID int64) ([]models.Service, error)
	GetServiceByID(ctx context.Context, serviceID int) (models.Service, error)
	GetServicesByStatus(ctx context.Context, status *models.ServiceStatus) ([]models.Service, error)
	UpdateService(ctx context.Context, service *models.Service) error
	SoftDeleteService(ctx context.Context, serviceID int) error
	UpdateServiceStatus(ctx context.Context, serviceID int, status models.ServiceStatus) error
}

// APIBoxService implements HTTP API logic for boxed solutions.
type APIBoxService struct {
	lister BoxLister
}

// NewAPIBoxService creates a new instance of the box service.
func NewAPIBoxService(lister BoxLister) *APIBoxService {
	return &APIBoxService{lister: lister}
}

// List returns all box solutions for API (telegramID=0 — все коробки).
func (s *APIBoxService) List(ctx context.Context) ([]dto.BoxListItem, error) {
	services, err := s.lister.GetServices(ctx, 0)
	if err != nil {
		return nil, err
	}
	out := make([]dto.BoxListItem, 0, len(services))
	for _, svc := range services {
		if svc.DeletedAt.Valid {
			continue
		}
		out = append(out, dto.BoxListItem{
			ID:          svc.ID,
			Name:        svc.Name,
			Description: svc.Description,
			Type:        svc.Type,
			BoxSolution: svc.BoxSolution,
		})
	}
	return out, nil
}

// GetByID returns a single box solution by ID.
func (s *APIBoxService) GetByID(ctx context.Context, id int) (*dto.BoxDetail, error) {
	svc, err := s.lister.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if svc.DeletedAt.Valid {
		return nil, nil
	}

	slots := make([]dto.BoxAvailableSlot, 0, len(svc.AvailableSlots))
	for _, s := range svc.AvailableSlots {
		slots = append(slots, dto.BoxAvailableSlot{
			Date:      s.Date,
			TimeSlots: s.TimeSlots,
		})
	}

	return &dto.BoxDetail{
		ID:             svc.ID,
		Name:           svc.Name,
		Description:    svc.Description,
		Rules:          svc.Rules,
		Schedule:       svc.Schedule,
		Type:           svc.Type,
		BoxSolution:    svc.BoxSolution,
		AvailableSlots: slots,
	}, nil
}

// Update updates the box
func (s *APIBoxService) Update(ctx context.Context, id int, req *dto.BoxUpdateRequest) (*dto.BoxDetail, error) {
	svc, err := s.lister.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if svc.DeletedAt.Valid {
		return nil, nil
	}

	if req.Name != nil {
		svc.Name = *req.Name
	}
	if req.Description != nil {
		svc.Description = req.Description
	}
	if req.Type != nil {
		svc.Type = *req.Type
	}
	if req.BoxSolution != nil {
		svc.BoxSolution = *req.BoxSolution
	}

	svc.UpdatedAt = time.Now()

	if err := s.lister.UpdateService(ctx, &svc); err != nil {
		return nil, err
	}

	slots := make([]dto.BoxAvailableSlot, 0, len(svc.AvailableSlots))
	for _, slot := range svc.AvailableSlots {
		slots = append(slots, dto.BoxAvailableSlot{
			Date:      slot.Date,
			TimeSlots: slot.TimeSlots,
		})
	}

	return &dto.BoxDetail{
		ID:             svc.ID,
		Name:           svc.Name,
		Description:    svc.Description,
		Rules:          svc.Rules,
		Schedule:       svc.Schedule,
		Type:           svc.Type,
		BoxSolution:    svc.BoxSolution,
		AvailableSlots: slots,
	}, nil
}

// Delete logical box deletion
func (s *APIBoxService) Delete(ctx context.Context, id int) error {
	svc, err := s.lister.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	if svc.DeletedAt.Valid {
		return nil
	}

	return s.lister.SoftDeleteService(ctx, id)

}

// UpdateStatus updates the status of the box
func (s *APIBoxService) UpdateStatus(ctx context.Context, id int, status models.ServiceStatus) (*dto.BoxStatusResponse, error) {
	if status != "active" && status != "hidden" && status != "draft" && status != "processed" {
		return nil, errors.New("invalid status")
	}

	svc, err := s.lister.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := s.lister.UpdateServiceStatus(ctx, id, status); err != nil {
		return nil, err
	}

	return &dto.BoxStatusResponse{
		ID:        id,
		Status:    string(status),
		UpdatedAt: time.Now(),
	}, nil
}

// Export exports the boxes in the specified format
func (s *APIBoxService) Export(ctx context.Context, status string, format string) ([]byte, string, error) {
	var statusEnum *models.ServiceStatus
	if status != "" {
		s := models.ServiceStatus(status)
		statusEnum = &s
	}

	services, err := s.lister.GetServicesByStatus(ctx, statusEnum)
	if err != nil {
		return nil, "", err
	}

	activeServices := make([]models.Service, 0)
	for _, svc := range services {
		if !svc.DeletedAt.Valid {
			activeServices = append(activeServices, svc)
		}
	}

	switch format {
	case "csv":
		data, err := s.generateCSV(activeServices)
		return data, "text/csv", err
	case "pdf":
		fallthrough
	default:
		data, err := s.generatePDF(activeServices)
		return data, "application/pdf", err
	}
}

// generateCSV creates a CSV file
func (s *APIBoxService) generateCSV(services []models.Service) ([]byte, error) {
	buf := &strings.Builder{}
	writer := csv.NewWriter(buf)

	if err := writer.Write([]string{"ID", "Name", "Description", "Type", "BoxSolution"}); err != nil {
		return nil, err
	}

	for _, svc := range services {
		if err := writer.Write([]string{
			strconv.Itoa(int(svc.ID)),
			svc.Name,
			svc.Description,
			svc.Type,
			strconv.FormatBool(svc.BoxSolution),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buf.String()), nil
}

// generatePDF creates a PDF file (stub)
func (s *APIBoxService) generatePDF(services []models.Service) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("Boxes Export\n\n")
	for _, svc := range services {
		buf.WriteString(svc.Name + "\n")
	}
	return []byte(buf.String()), nil
}
