package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

// BoxSolutionRepo the repository of boxed solutions
type BoxSolutionRepo struct {
	db *sqlx.DB
}

// NewBoxSolutionRepo returns a new instance of the boxed solutions repository
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

const getAvailableTimeSlotsQuery = `
	SELECT 
		slot_date,
		time_slots
	FROM service_available_slots 
	WHERE service_id = $1
	ORDER BY slot_date`

const getTimeSlotsByDateQuery = `
	SELECT time_slots
	FROM service_available_slots 
	WHERE service_id = $1 AND slot_date = $2`

const checkSlotAvailabilityQuery = `
	SELECT EXISTS(
		SELECT 1 
		FROM service_available_slots 
		WHERE service_id = $1 
			AND slot_date = $2 
			AND $3 = ANY(time_slots)
	)`

// GetServices gets a list of all services marked as boxed solutions
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

// GetServiceByID gets detailed information about a specific service by its identifier
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

// GetAvailableDates gets a list of available dates for the service
func (r *BoxSolutionRepo) GetAvailableDates(ctx context.Context, serviceID int) ([]string, error) {
	if serviceID <= 0 {
		return nil, errors.New("invalid service ID")
	}

	type dbSlot struct {
		SlotDate sql.NullTime `db:"slot_date"`
	}

	query := `
		SELECT slot_date
		FROM service_available_slots 
		WHERE service_id = $1 AND array_length(time_slots, 1) > 0
		ORDER BY slot_date
	`

	var dbSlots []dbSlot
	err := r.db.SelectContext(ctx, &dbSlots, query, serviceID)
	if err != nil {
		logger.Error("failed to get available dates from db",
			zap.Int("service_id", serviceID),
			zap.Error(err))
		return nil, err
	}

	dates := make([]string, 0, len(dbSlots))
	for _, slot := range dbSlots {
		if slot.SlotDate.Valid {
			dates = append(dates, slot.SlotDate.Time.Format("2006-01-02"))
		}
	}

	return dates, nil
}

// GetAvailableSlotsByServiceID gets all available dates and time slots for the service
func (r *BoxSolutionRepo) GetAvailableSlotsByServiceID(ctx context.Context, serviceID int) ([]models.AvailableSlot, error) {
	if serviceID <= 0 {
		return nil, errors.New("invalid service ID")
	}

	type dbSlot struct {
		SlotDate  sql.NullTime   `db:"slot_date"`
		TimeSlots pq.StringArray `db:"time_slots"`
	}

	var dbSlots []dbSlot
	err := r.db.SelectContext(ctx, &dbSlots, getAvailableTimeSlotsQuery, serviceID)
	if err != nil {
		logger.Error("failed to get available slots from db",
			zap.Int("service_id", serviceID),
			zap.Error(err))
		return nil, err
	}

	availableSlots := make([]models.AvailableSlot, 0, len(dbSlots))
	for _, slot := range dbSlots {
		if slot.SlotDate.Valid && len(slot.TimeSlots) > 0 {
			availableSlots = append(availableSlots, models.AvailableSlot{
				Date:      slot.SlotDate.Time.Format("2006-01-02"),
				TimeSlots: slot.TimeSlots,
			})
		}
	}

	return availableSlots, nil
}

// GetAvailableTimeSlotsByDate gets available time slots for a specific date
func (r *BoxSolutionRepo) GetAvailableTimeSlotsByDate(ctx context.Context, serviceID int, date string) ([]string, error) {
	if serviceID <= 0 {
		return nil, errors.New("invalid service ID")
	}
	if date == "" {
		return nil, errors.New("date cannot be empty")
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}

	var timeSlots pq.StringArray
	err := r.db.GetContext(ctx, &timeSlots, getTimeSlotsByDateQuery, serviceID, date)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		logger.Error("failed to get time slots from db",
			zap.Int("service_id", serviceID),
			zap.String("date", date),
			zap.Error(err))
		return nil, err
	}

	return timeSlots, nil
}

// CheckSlotAvailability checks the availability of a specific time slot
func (r *BoxSolutionRepo) CheckSlotAvailability(ctx context.Context, serviceID int, date string, timeSlot string) (bool, error) {
	if serviceID <= 0 {
		return false, errors.New("invalid service ID")
	}
	if date == "" {
		return false, errors.New("date cannot be empty")
	}
	if timeSlot == "" {
		return false, errors.New("time slot cannot be empty")
	}

	if _, err := time.Parse("2006-01-02", date); err != nil {
		return false, errors.New("invalid date format, expected YYYY-MM-DD")
	}

	var exists bool
	err := r.db.GetContext(ctx, &exists, checkSlotAvailabilityQuery, serviceID, date, timeSlot)
	if err != nil {
		logger.Error("failed to check slot availability",
			zap.Int("service_id", serviceID),
			zap.String("date", date),
			zap.String("time_slot", timeSlot),
			zap.Error(err))
		return false, err
	}

	return exists, nil
}
