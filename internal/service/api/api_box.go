package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

// BoxLister returns box solutions (services with box_solution=true) from storage.
//

//go:generate mockgen -source=../../repository/interfaces.go -destination=mocks/mock_box_solution_repository.go -package=mocks
// type BoxLister interface {
// 	GetServices(ctx context.Context, telegramID int64) (*[]models.Service, error)
// 	GetServiceByID(ctx context.Context, serviceID int64) (*models.Service, error)
// 	// GetServicesByStatus(ctx context.Context, status *models.ServiceStatus) ([]models.Service, error)
// 	UpdateService(ctx context.Context, id int64, service *models.BoxUpdate) error
// 	SoftDeleteService(ctx context.Context, serviceID int64) error
// 	UpdateServiceStatus(ctx context.Context, serviceID int64, status models.ServiceStatus) (*models.BoxUpdateStatusResult, error)
// 	UpdateServiceSlots(ctx context.Context, id int64, slots *models.BoxNewSlots) error
// 	DeleteServiceSlots(ctx context.Context, id int64) error
// }

// APIBoxService implements HTTP API logic for boxed solutions.
type APIBoxService struct {
	lister      repository.BoxSolutionRepository
	fileService *FileService
	txRepo      repository.TxRepository
}

// NewAPIBoxService creates a new instance of the box service.
func NewAPIBoxService(lister repository.BoxSolutionRepository,
	fileService *FileService,
	txRepo repository.TxRepository) *APIBoxService {
	return &APIBoxService{
		lister:      lister,
		fileService: fileService,
		txRepo:      txRepo,
	}
}

// List returns all box solutions for API
func (s *APIBoxService) List(ctx context.Context, query models.BoxList) (*models.BoxListResult, error) {
	result, err := s.lister.List(ctx, query)
	if err != nil {
		logger.Error("failed to get boxes list",
			zap.Error(err),
			zap.Any("query", query))
		return nil, models.ErrDatabase
	}

	if result == nil {
		return &models.BoxListResult{
			Items:  []models.Service{},
			Total:  0,
			Limit:  query.Limit,
			Offset: query.Offset,
		}, nil
	}

	return result, nil
}

// Create creates a box
func (s *APIBoxService) Create(ctx context.Context, box *models.BoxCreate) (*models.Service, error) {
	service, err := s.lister.CreateBox(ctx, box)
	if err != nil {
		return service, err
	}
	return service, nil
}

// GetByID returns a single box solution by ID.
func (s *APIBoxService) GetByID(ctx context.Context, id int64) (*models.Service, error) {
	svc, err := s.lister.GetServiceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// Update updates the box
func (s *APIBoxService) Update(ctx context.Context, id int64, req *models.BoxUpdate) (*models.Service, error) {
	boxNewSlots, err := parseBoxSlots(req.Slots)
	if err != nil {
		return nil, err
	}

	var svc *models.Service
	err = s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {
		err := s.lister.UpdateService(txCtx, id, req)
		if err != nil {
			return err
		}

		// nil = slots unchanged, [] = delete all slots, [...] = replace
		if req.Slots != nil {
			if err = s.lister.DeleteServiceSlots(txCtx, id); err != nil {
				return err
			}
		}

		if len(req.Slots) > 0 {
			err = s.lister.UpdateServiceSlots(txCtx, id, &boxNewSlots)
		}
		if err != nil {
			return err
		}

		svc, err = s.lister.GetServiceByID(txCtx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	return svc, nil
}

// Delete logical box deletion
func (s *APIBoxService) Delete(ctx context.Context, id int64) error {
	return s.lister.SoftDeleteService(ctx, id)
}

// UpdateStatus updates the status of the box
func (s *APIBoxService) UpdateStatus(ctx context.Context, id int64, status models.ServiceStatus) (*models.BoxUpdateStatusResult, error) {

	res, err := s.lister.UpdateServiceStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	return res, nil
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

	activeServices = append(activeServices, services...)

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

func parseBoxSlots(slots []models.BoxAvailableSlot) (models.BoxNewSlots, error) {
	var boxNewSlots models.BoxNewSlots
	if len(slots) > 0 {
		boxNewSlots = models.BoxNewSlots{
			Date:      make([]time.Time, len(slots)),
			StartTime: make([]time.Time, len(slots)),
			EndTime:   make([]time.Time, len(slots)),
		}
		for i, t := range slots {
			date, err := time.Parse("2006-01-02", t.Date)
			if err != nil {
				return models.BoxNewSlots{}, models.ErrInvalidInput
			}
			startTime, err := time.Parse("15:04", t.StartTime)
			if err != nil {
				return models.BoxNewSlots{}, models.ErrInvalidInput
			}
			endTime, err := time.Parse("15:04", t.EndTime)
			if err != nil {
				return models.BoxNewSlots{}, models.ErrInvalidInput
			}
			boxNewSlots.Date[i] = date
			boxNewSlots.StartTime[i] = startTime
			boxNewSlots.EndTime[i] = endTime
		}
	}

	return boxNewSlots, nil
}
