package repository

import (
	"context"
	"database/sql"
	"time"

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

func (r *BoxSolutionRepo) GetServicesByStatus(ctx context.Context, status *models.ServiceStatus) ([]models.Service, error) {
	query := `
        SELECT
            s.id, s.name, s.description, s.rules, s.schedule,
            s.type, s.box_solution, s.status, s.created_at, s.updated_at, s.deleted_at,
            a.slot_date, COALESCE(a.time_slots, '{}') as time_slots
        FROM services s
        LEFT JOIN service_available_slots a ON s.id = a.service_id
        WHERE s.box_solution = true AND s.deleted_at IS NULL`

	var args []interface{}
	if status != nil {
		query += ` AND s.status = $1`
		args = append(args, *status)
	}
	query += ` ORDER BY s.id, a.slot_date`

	type serviceRow struct {
		ID          int64                `db:"id"`
		Name        string               `db:"name"`
		Description sql.NullString       `db:"description"`
		Rules       sql.NullString       `db:"rules"`
		Schedule    sql.NullString       `db:"schedule"`
		Type        sql.NullString       `db:"type"`
		BoxSolution bool                 `db:"box_solution"`
		Status      models.ServiceStatus `db:"status"`
		CreatedAt   time.Time            `db:"created_at"`
		UpdatedAt   time.Time            `db:"updated_at"`
		DeletedAt   sql.NullTime         `db:"deleted_at"`
		SlotDate    sql.NullTime         `db:"slot_date"`
		TimeSlots   pq.StringArray       `db:"time_slots"`
	}

	var rows []serviceRow
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	serviceMap := make(map[int64]*models.Service)
	for _, row := range rows {
		svc, exists := serviceMap[row.ID]
		if !exists {
			svc = &models.Service{
				ID:             row.ID,
				Name:           row.Name,
				Description:    row.Description.String,
				Rules:          row.Rules.String,
				Schedule:       row.Schedule.String,
				Type:           row.Type.String,
				BoxSolution:    row.BoxSolution,
				Status:         row.Status,
				CreatedAt:      row.CreatedAt,
				UpdatedAt:      row.UpdatedAt,
				DeletedAt:      row.DeletedAt,
				AvailableSlots: []models.AvailableSlot{},
			}
			serviceMap[row.ID] = svc
		}
		if row.SlotDate.Valid {
			svc.AvailableSlots = append(svc.AvailableSlots, models.AvailableSlot{
				Date:      row.SlotDate.Time.Format("2006-01-02"),
				TimeSlots: row.TimeSlots,
			})
		}
	}

	result := make([]models.Service, 0, len(serviceMap))
	for _, svc := range serviceMap {
		result = append(result, *svc)
	}
	return result, nil
}

func (r *BoxSolutionRepo) UpdateService(ctx context.Context, service *models.Service) error {
	query := `
        UPDATE services 
        SET name = $2, description = $3, rules = $4, schedule = $5, 
            type = $6, box_solution = $7, updated_at = $8
        WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		service.ID, service.Name, service.Description, service.Rules,
		service.Schedule, service.Type, service.BoxSolution, service.UpdatedAt)
	return err
}

func (r *BoxSolutionRepo) SoftDeleteService(ctx context.Context, serviceID int) error {
	query := `UPDATE services SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, serviceID)
	return err
}

func (r *BoxSolutionRepo) UpdateServiceStatus(ctx context.Context, serviceID int, status models.ServiceStatus) error {
	query := `UPDATE services SET status = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, serviceID, status)
	return err
}
