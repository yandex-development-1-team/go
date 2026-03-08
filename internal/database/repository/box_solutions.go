package repository

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

const getServicesQuery = `
SELECT
    s.id,
    s.name,
    s.description,
    s.rules,
    s.schedule,
    s.type,
    s.box_solution,
    a.slot_date,
    COALESCE(a.time_slots, '{}') AS time_slots
FROM services s
LEFT JOIN service_available_slots a ON s.id = a.service_id
WHERE s.box_solution = true
ORDER BY s.id, a.slot_date`

func (r Repository) GetServices(ctx context.Context, telegramID int64) ([]models.Service, error) {

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

	return withMetricsValue("get_services", func() ([]models.Service, error) {
		var bsServices []service
		if err := r.client.SelectContext(ctx, &bsServices, getServicesQuery); err != nil {
			logger.Error("failed to get box solutions from db", zap.Int64("chat_id", telegramID), zap.Error(err))
			return nil, err
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
	})
}
