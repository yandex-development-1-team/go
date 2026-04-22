package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/dto"
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
		s.id, s.name, s.slug, s.description, s.rules, s.location, s.price, s.image,
		s.status, s.organizer, s.created_at, s.updated_at
	FROM services s
	WHERE s.deleted_at IS NULL AND s.status = 'active'
	ORDER BY s.id`

const getServiceByIDQuery = `
	SELECT
		s.id, s.name, s.slug, s.description, s.rules, s.location, s.price, s.image,
		s.status, s.organizer, s.created_at, s.updated_at,
		a.slot_date, a.start_time, a.end_time
	FROM services s
	LEFT JOIN service_available_slots a ON s.id = a.service_id
	WHERE s.id = $1 AND s.deleted_at IS NULL
	ORDER BY a.slot_date, a.start_time`

const updateServiceByIDQuery = `
	UPDATE services SET
		name        = COALESCE($2, name),
		description = COALESCE($3, description),
		rules       = COALESCE($4, rules),
		location    = COALESCE($5, location),
		price       = COALESCE($6, price),
		image       = COALESCE($7, image),
		status      = COALESCE($8, status),
		organizer   = COALESCE($9, organizer),
		updated_at  = NOW()
	WHERE id = $1`

const createServiceQuery = `
	INSERT INTO services (
		name, slug, description, rules, location, price, image, status, organizer
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at, updated_at`

const createAvailableSlotQuery = `
	INSERT INTO service_available_slots (
		service_id, slot_date, start_time, end_time
	) VALUES ($1, $2, $3::time, $4::time)`

const deleteSlotsQuery = `
	DELETE FROM service_available_slots
		WHERE service_id=$1`

const createSlotsQuery = `
	INSERT INTO service_available_slots
		(service_id, slot_date, start_time, end_time)
		SELECT $1, unnest($2::date[]), unnest($3::time[]),
		unnest($4::time[])`

const updateStatusQuery = `
	UPDATE services 
		SET status = $2 
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, status, updated_at`

// List returns a list of boxes with filtering, pagination and sorting
func (r *BoxSolutionRepo) List(ctx context.Context, query models.BoxList) (*models.BoxListResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.Sort == "" {
		query.Sort = "created_at"
	}
	if query.Order == "" {
		query.Order = "asc"
	}

	validSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"created_at": true,
		"updated_at": true,
	}
	if !validSortFields[query.Sort] {
		query.Sort = "created_at"
	}

	if query.Order != "asc" && query.Order != "desc" {
		query.Order = "asc"
	}

	where := []string{"s.deleted_at IS NULL"}
	args := []interface{}{}
	argPos := 1

	if query.Status != nil && *query.Status != "" {
		where = append(where, fmt.Sprintf("s.status = $%d", argPos))
		args = append(args, *query.Status)
		argPos++
	}

	if query.Search != nil && *query.Search != "" {
		where = append(where, fmt.Sprintf("s.name ILIKE $%d", argPos))
		args = append(args, "%"+*query.Search+"%")
		argPos++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM services s
		WHERE %s
	`, whereClause)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Error("failed to count boxes",
				zap.Error(err),
				zap.String("query", countQuery),
				zap.Any("args", args))
		}
		return nil, fmt.Errorf("count boxes: %w", err)
	}

	if total == 0 {
		return &models.BoxListResult{
			Items:  []models.Service{},
			Total:  0,
			Limit:  query.Limit,
			Offset: query.Offset,
		}, nil
	}

	orderBy := fmt.Sprintf("%s %s", query.Sort, query.Order)

	dataQuery := fmt.Sprintf(`
		SELECT
			s.id, s.name, s.slug, s.description, s.rules, s.location, s.price, s.image,
			s.status, s.organizer, s.created_at, s.updated_at
		FROM services s
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argPos, argPos+1)

	dataArgs := append(args, query.Limit, query.Offset)

	type ServiceRaw struct {
		ID          int64          `db:"id"`
		Name        string         `db:"name"`
		Slug        string         `db:"slug"`
		Description sql.NullString `db:"description"`
		Rules       sql.NullString `db:"rules"`
		Location    sql.NullString `db:"location"`
		Price       int            `db:"price"`
		Image       *string        `db:"image"`
		Status      string         `db:"status"`
		Organizer   sql.NullString `db:"organizer"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}

	var serviceRows []ServiceRaw
	err = r.db.SelectContext(ctx, &serviceRows, dataQuery, dataArgs...)
	if err != nil {
		logger.Error("failed to list boxes",
			zap.Error(err),
			zap.String("query", dataQuery),
			zap.Any("args", dataArgs))
		return nil, fmt.Errorf("list boxes: %w", err)
	}

	if len(serviceRows) == 0 {
		return &models.BoxListResult{
			Items:  []models.Service{},
			Total:  total,
			Limit:  query.Limit,
			Offset: query.Offset,
		}, nil
	}

	serviceIDs := make([]int64, len(serviceRows))
	for i, row := range serviceRows {
		serviceIDs[i] = row.ID
	}

	slotQuery := `
		SELECT service_id, slot_date, start_time, end_time
		FROM service_available_slots
		WHERE service_id = ANY($1)
		ORDER BY service_id, slot_date, start_time
	`

	var slots []struct {
		ServiceID int64        `db:"service_id"`
		SlotDate  sql.NullTime `db:"slot_date"`
		StartTime sql.NullTime `db:"start_time"`
		EndTime   sql.NullTime `db:"end_time"`
	}

	err = r.db.SelectContext(ctx, &slots, slotQuery, serviceIDs)
	if err != nil {
		logger.Error("failed to get slots", zap.Error(err))
	}

	slotsMap := make(map[int64][]models.BoxAvailableSlot)
	for _, slot := range slots {
		if slot.SlotDate.Valid {
			slotsMap[slot.ServiceID] = append(slotsMap[slot.ServiceID], models.BoxAvailableSlot{
				Date:      slot.SlotDate.Time.Format("2006-01-02"),
				StartTime: slot.StartTime.Time.Format("15:04"),
				EndTime:   slot.EndTime.Time.Format("15:04"),
			})
		}
	}

	items := make([]models.Service, 0, len(serviceRows))
	for _, row := range serviceRows {
		svc := models.Service{
			ID:                row.ID,
			Name:              row.Name,
			Slug:              row.Slug,
			Description:       nullStringToString(row.Description),
			Rules:             nullStringToString(row.Rules),
			Location:          nullStringToString(row.Location),
			Price:             row.Price,
			Image:             row.Image,
			Status:            row.Status,
			Organizer:         nullStringToString(row.Organizer),
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			BoxAvailableSlots: slotsMap[row.ID],
		}
		items = append(items, svc)
	}

	return &models.BoxListResult{
		Items:  items,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, nil
}

// nullStringToString converts sql.NullString to string
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// GetServices gets a list of all active services (boxed solutions)
func (r *BoxSolutionRepo) GetServices(ctx context.Context, telegramID int64) ([]models.Service, error) {
	type ServiceRaw struct {
		ID          int64          `db:"id"`
		Name        string         `db:"name"`
		Slug        string         `db:"slug"`
		Description sql.NullString `db:"description"`
		Rules       sql.NullString `db:"rules"`
		Location    sql.NullString `db:"location"`
		Price       int            `db:"price"`
		Image       *string        `db:"image"`
		Status      string         `db:"status"`
		Organizer   sql.NullString `db:"organizer"`
		CreatedAt   time.Time      `db:"created_at"`
		UpdatedAt   time.Time      `db:"updated_at"`
	}

	var serviceRows []ServiceRaw
	err := r.db.SelectContext(ctx, &serviceRows, getBoxServicesQuery)
	if err != nil {
		logger.Error("failed to get services from db", zap.Int64("chat_id", telegramID), zap.Error(err))
		return nil, err
	}

	if len(serviceRows) == 0 {
		return []models.Service{}, nil
	}

	serviceIDs := make([]int64, len(serviceRows))
	for i, row := range serviceRows {
		serviceIDs[i] = row.ID
	}

	slotQuery := `
        SELECT service_id, slot_date, start_time, end_time
        FROM service_available_slots
        WHERE service_id = ANY($1)
        ORDER BY service_id, slot_date, start_time
    `

	var slots []struct {
		ServiceID int64        `db:"service_id"`
		SlotDate  sql.NullTime `db:"slot_date"`
		StartTime sql.NullTime `db:"start_time"`
		EndTime   sql.NullTime `db:"end_time"`
	}

	err = r.db.SelectContext(ctx, &slots, slotQuery, serviceIDs)
	if err != nil {
		logger.Error("failed to get slots for services", zap.Int64("chat_id", telegramID), zap.Error(err))
	}

	slotsMap := make(map[int64][]models.BoxAvailableSlot)
	for _, slot := range slots {
		if slot.SlotDate.Valid {
			slotsMap[slot.ServiceID] = append(slotsMap[slot.ServiceID], models.BoxAvailableSlot{
				Date:      slot.SlotDate.Time.Format("2006-01-02"),
				StartTime: slot.StartTime.Time.Format("15:04"),
				EndTime:   slot.EndTime.Time.Format("15:04"),
			})
		}
	}

	items := make([]models.Service, 0, len(serviceRows))
	for _, row := range serviceRows {
		svc := models.Service{
			ID:                row.ID,
			Name:              row.Name,
			Slug:              row.Slug,
			Description:       nullStringToString(row.Description),
			Rules:             nullStringToString(row.Rules),
			Location:          nullStringToString(row.Location),
			Price:             row.Price,
			Image:             row.Image,
			Status:            row.Status,
			Organizer:         nullStringToString(row.Organizer),
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			BoxAvailableSlots: slotsMap[row.ID],
		}
		items = append(items, svc)
	}

	return items, nil
}

// GetServiceByID gets detailed information about a specific service by its identifier
func (r *BoxSolutionRepo) GetServiceByID(ctx context.Context, serviceID int64) (*models.Service, error) {
	var rows []dto.BoxRaw
	if err := r.db.SelectContext(ctx, &rows, getServiceByIDQuery, serviceID); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, models.ErrBoxSolutionNotFound
	}

	svc := &models.Service{
		ID:          rows[0].ID,
		Name:        rows[0].Name,
		Slug:        rows[0].Slug,
		Description: derefString(rows[0].Description),
		Rules:       derefString(rows[0].Rules),
		Location:    derefString(rows[0].Location),
		Price:       rows[0].Price,
		Image:       rows[0].Image,
		Status:      string(rows[0].Status),
		Organizer:   derefString(rows[0].Organizer),
		CreatedAt:   rows[0].CreatedAt,
		UpdatedAt:   rows[0].UpdatedAt,
	}

	for _, row := range rows {
		if row.SlotDate.Valid {
			svc.BoxAvailableSlots = append(svc.BoxAvailableSlots, models.BoxAvailableSlot{
				Date:      row.SlotDate.Time.Format("2006-01-02"),
				StartTime: row.StartTime.Time.Format("15:04"),
				EndTime:   row.EndTime.Time.Format("15:04"),
			})
		}
	}

	return svc, nil
}

// GetAvailableSlotsByServiceID gets all available slots for the service
func (r *BoxSolutionRepo) GetAvailableSlotsByServiceID(ctx context.Context, serviceID int64) ([]models.BoxAvailableSlot, error) {
	if serviceID <= 0 {
		return nil, errors.New("invalid service ID")
	}

	type dbSlot struct {
		SlotDate  sql.NullTime `db:"slot_date"`
		StartTime sql.NullTime `db:"start_time"`
		EndTime   sql.NullTime `db:"end_time"`
	}

	query := `
		SELECT slot_date, start_time, end_time
		FROM service_available_slots
		WHERE service_id = $1
		ORDER BY slot_date, start_time
	`

	var dbSlots []dbSlot
	err := r.db.SelectContext(ctx, &dbSlots, query, serviceID)
	if err != nil {
		logger.Error("failed to get available slots from db",
			zap.Int64("service_id", serviceID),
			zap.Error(err))
		return nil, err
	}

	availableSlots := make([]models.BoxAvailableSlot, 0, len(dbSlots))
	for _, slot := range dbSlots {
		if slot.SlotDate.Valid {
			availableSlot := models.BoxAvailableSlot{
				Date: slot.SlotDate.Time.UTC().Format("2006-01-02"),
			}
			if slot.StartTime.Valid {
				availableSlot.StartTime = slot.StartTime.Time.UTC().Format("15:04")
			}
			if slot.EndTime.Valid {
				availableSlot.EndTime = slot.EndTime.Time.UTC().Format("15:04")
			}
			availableSlots = append(availableSlots, availableSlot)
		}
	}

	return availableSlots, nil
}

// CheckSlotAvailability checks the availability of a specific slot
func (r *BoxSolutionRepo) CheckSlotAvailability(ctx context.Context, serviceID int64, slot models.BoxAvailableSlot) (bool, error) {
	if serviceID <= 0 {
		return false, errors.New("invalid service ID")
	}
	if slot.Date == "" {
		return false, errors.New("date cannot be empty")
	}

	logger.Info("CheckSlotAvailability",
		zap.Int64("service_id", serviceID),
		zap.String("date", slot.Date),
		zap.String("start_time", slot.StartTime),
		zap.String("end_time", slot.EndTime))

	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM service_available_slots 
			WHERE service_id = $1 
				AND slot_date = $2 
				AND start_time = $3::time
				AND end_time = $4::time
		)`

	err := r.db.QueryRowContext(ctx, query, serviceID, slot.Date, slot.StartTime, slot.EndTime).Scan(&exists)
	if err != nil {
		logger.Error("failed to check slot availability", zap.Error(err))
		return false, err
	}

	logger.Info("CheckSlotAvailability result", zap.Bool("exists", exists))
	return exists, nil
}

// CreateBox creates a new boxed solution with its available slots
func (r *BoxSolutionRepo) CreateBox(ctx context.Context, box *models.BoxCreate) (*models.Service, error) {
	if box == nil {
		return nil, errors.New("service cannot be nil")
	}

	if box.Name == nil || *box.Name == "" {
		return nil, errors.New("name is required")
	}
	if box.Slug == nil || *box.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if box.Price == nil {
		return nil, errors.New("price is required")
	}
	if box.Status == nil || *box.Status == "" {
		return nil, errors.New("status is required")
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var id int64
	var createdAt, updatedAt time.Time

	name := *box.Name
	slug := *box.Slug
	price := *box.Price
	status := *box.Status

	var description interface{} = nil
	if box.Description != nil && *box.Description != "" {
		description = *box.Description
	}
	var rules interface{} = nil
	if box.Rules != nil && *box.Rules != "" {
		rules = *box.Rules
	}
	var location interface{} = nil
	if box.Location != nil && *box.Location != "" {
		location = *box.Location
	}
	var image interface{} = nil
	if box.Image != nil && *box.Image != "" {
		image = *box.Image
	}
	var organizer interface{} = nil
	if box.Organizer != nil && *box.Organizer != "" {
		organizer = *box.Organizer
	}

	err = tx.QueryRowContext(ctx, createServiceQuery,
		name,
		slug,
		description,
		rules,
		location,
		price,
		image,
		status,
		organizer,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	service := &models.Service{
		ID:                id,
		Name:              name,
		Slug:              slug,
		Description:       getStringValue(box.Description),
		Rules:             getStringValue(box.Rules),
		Location:          getStringValue(box.Location),
		Price:             price,
		Image:             box.Image,
		Status:            status,
		Organizer:         getStringValue(box.Organizer),
		BoxAvailableSlots: box.Slots,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}

	if len(service.BoxAvailableSlots) > 0 {
		for _, slot := range service.BoxAvailableSlots {
			if slot.Date == "" {
				return nil, fmt.Errorf("slot date cannot be empty")
			}

			parsedDate, err := time.ParseInLocation("2006-01-02", slot.Date, time.UTC)
			if err != nil {
				return nil, fmt.Errorf("invalid date format: %w", err)
			}

			var startTime interface{} = nil
			var endTime interface{} = nil

			if slot.StartTime != "" {
				startTime = slot.StartTime
			}
			if slot.EndTime != "" {
				endTime = slot.EndTime
			}

			_, err = tx.ExecContext(ctx, createAvailableSlotQuery,
				service.ID,
				parsedDate,
				startTime,
				endTime,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create available slot: %w", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return service, nil
}

// getStringValue returns the value of a string or an empty string
func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func (r *BoxSolutionRepo) GetServicesByStatus(ctx context.Context, status *models.ServiceStatus) ([]models.Service, error) {
	query := `
		SELECT
			s.id, s.name, s.slug, s.description, s.rules, s.location,
			s.price, s.image, s.status, s.organizer, s.created_at, s.updated_at,
			a.slot_date, a.start_time, a.end_time
		FROM services s
		LEFT JOIN service_available_slots a ON s.id = a.service_id
		WHERE s.deleted_at IS NULL`

	var args []interface{}
	if status != nil {
		query += ` AND s.status = $1`
		args = append(args, *status)
	}
	query += ` ORDER BY s.id, a.slot_date, a.start_time`

	var rows []dto.BoxRaw

	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}

	serviceMap := make(map[int64]*models.Service)
	for _, row := range rows {
		svc, exists := serviceMap[row.ID]
		if !exists {
			svc = &models.Service{
				ID:                row.ID,
				Name:              row.Name,
				Slug:              row.Slug,
				Description:       derefString(row.Description),
				Rules:             derefString(row.Rules),
				Location:          derefString(row.Location),
				Price:             row.Price,
				Image:             row.Image,
				Status:            string(row.Status),
				Organizer:         derefString(row.Organizer),
				CreatedAt:         row.CreatedAt,
				UpdatedAt:         row.UpdatedAt,
				BoxAvailableSlots: []models.BoxAvailableSlot{},
			}
			serviceMap[row.ID] = svc
		}
		if row.SlotDate.Valid {
			svc.BoxAvailableSlots = append(svc.BoxAvailableSlots, models.BoxAvailableSlot{
				Date:      row.SlotDate.Time.Format("2006-01-02"),
				StartTime: row.StartTime.Time.Format("15:04"),
				EndTime:   row.EndTime.Time.Format("15:04"),
			})
		}
	}

	result := make([]models.Service, 0, len(serviceMap))
	for _, svc := range serviceMap {
		result = append(result, *svc)
	}
	return result, nil
}

func (r *BoxSolutionRepo) UpdateService(ctx context.Context, id int64, service *models.BoxUpdate) error {
	result, err := r.db.ExecContext(ctx, updateServiceByIDQuery,
		id, service.Name, service.Description, service.Rules, service.Location,
		service.Price, service.Image, service.Status, service.Organizer)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return models.ErrBoxSolutionNotFound
	}

	return nil
}

func (r *BoxSolutionRepo) DeleteServiceSlots(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, deleteSlotsQuery, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *BoxSolutionRepo) UpdateServiceSlots(ctx context.Context, id int64, slots *models.BoxNewSlots) error {

	_, err := r.db.ExecContext(ctx, createSlotsQuery, id, slots.Date, slots.StartTime, slots.EndTime)
	if err != nil {
		return err
	}

	return nil
}

func (r *BoxSolutionRepo) SoftDeleteService(ctx context.Context, serviceID int64) error {
	query := `UPDATE services SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return err
	}

	return nil
}

func (r *BoxSolutionRepo) UpdateServiceStatus(ctx context.Context, serviceID int64, status models.ServiceStatus) (*models.BoxUpdateStatusResult, error) {
	var result dto.BoxUpdateStatusResult
	err := r.db.GetContext(ctx, &result, updateStatusQuery, serviceID, status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrBoxSolutionNotFound // изменил 30,03,2026
	}
	if err != nil {
		return nil, err
	}

	return &models.BoxUpdateStatusResult{
		ID:        result.ID,
		Status:    result.Status,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
