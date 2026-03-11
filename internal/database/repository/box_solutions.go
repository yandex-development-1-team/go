package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

// BoxSolutionRepo — репозиторий коробочных решений (services с box_solution=true).
type BoxSolutionRepo struct {
	db *sqlx.DB
}

func NewBoxSolutionRepo(db *sqlx.DB) *BoxSolutionRepo {
	return &BoxSolutionRepo{db: db}
}

const getBoxServicesQuery = `
	SELECT
		s.id, s.name, s.description, s.rules, s.schedule,
		s.type, s.box_solution, a.slot_date,
		COALESCE(a.time_slots, '{}') as time_slots
	FROM services s
	LEFT JOIN service_available_slots a ON s.id = a.service_id
	WHERE s.box_solution = true
	ORDER BY s.id, a.slot_date`

const getServiceByIDQuery = `
	SELECT
		s.id, s.name, s.description, s.rules, s.schedule,
		s.type, s.box_solution, a.slot_date,
		COALESCE(a.time_slots, '{}') as time_slots
	FROM services s
	LEFT JOIN service_available_slots a ON s.id = a.service_id
	WHERE s.id = $1
	ORDER BY a.slot_date`

func (r *BoxSolutionRepo) GetServices(ctx context.Context, telegramID int64) ([]models.Service, error) {
	type service struct {
		ID          int64          `db:"id"`
		Name        string         `db:"name"`
		Description sql.NullString `db:"description"`
		Rules       sql.NullString `db:"rules"`
		Schedule    sql.NullString `db:"schedule"`
		Type        sql.NullString `db:"type"`
		BoxSolution bool           `db:"box_solution"`
		SlotDate    sql.NullTime   `db:"slot_date"`
		TimeSlots   pq.StringArray `db:"time_slots"`
	}

	var bsServices []service
	var boxSolutionServices []models.Service

	err := r.db.SelectContext(ctx, &bsServices, getBoxServicesQuery)
	if err != nil {
		logger.Error("failed to get box solutions from db", zap.Int64("chat_id", telegramID), zap.Error(err))
		return boxSolutionServices, err
	}

	bsServicesMap := make(map[int64]*models.Service)

	for _, bsService := range bsServices {
		boxSolutionService, exists := bsServicesMap[bsService.ID]
		if !exists {
			boxSolutionService = &models.Service{
				ID:             bsService.ID,
				Name:           bsService.Name,
				Description:    bsService.Description.String,
				Rules:          bsService.Rules.String,
				Schedule:       bsService.Schedule.String,
				Type:           bsService.Type.String,
				BoxSolution:    bsService.BoxSolution,
				AvailableSlots: []models.AvailableSlot{},
			}
			bsServicesMap[bsService.ID] = boxSolutionService
		}

		if bsService.SlotDate.Valid {
			boxSolutionService.AvailableSlots = append(boxSolutionService.AvailableSlots, models.AvailableSlot{
				Date:      bsService.SlotDate.Time.Format("2006-01-02"),
				TimeSlots: bsService.TimeSlots,
			})
		}
	}

	services := make([]models.Service, 0, len(bsServicesMap))
	for _, boxSolutionService := range bsServicesMap {
		services = append(services, *boxSolutionService)
	}

	return services, nil
}

func (r *BoxSolutionRepo) GetServiceByID(ctx context.Context, serviceID int) (models.Service, error) {
	type row struct {
		ID          int64          `db:"id"`
		Name        string         `db:"name"`
		Description sql.NullString `db:"description"`
		Rules       sql.NullString `db:"rules"`
		Schedule    sql.NullString `db:"schedule"`
		Type        sql.NullString `db:"type"`
		BoxSolution bool           `db:"box_solution"`
		SlotDate    sql.NullTime   `db:"slot_date"`
		TimeSlots   pq.StringArray `db:"time_slots"`
	}

	var rows []row
	if err := r.db.SelectContext(ctx, &rows, getServiceByIDQuery, serviceID); err != nil {
		return models.Service{}, err
	}
	if len(rows) == 0 {
		return models.Service{}, sql.ErrNoRows
	}

	svc := models.Service{
		ID:             rows[0].ID,
		Name:           rows[0].Name,
		Description:    rows[0].Description.String,
		Rules:          rows[0].Rules.String,
		Schedule:       rows[0].Schedule.String,
		Type:           rows[0].Type.String,
		BoxSolution:    rows[0].BoxSolution,
		AvailableSlots: make([]models.AvailableSlot, 0),
	}
	for _, r := range rows {
		if r.SlotDate.Valid {
			svc.AvailableSlots = append(svc.AvailableSlots, models.AvailableSlot{
				Date:      r.SlotDate.Time.Format("2006-01-02"),
				TimeSlots: r.TimeSlots,
			})
		}
	}
	return svc, nil
}
