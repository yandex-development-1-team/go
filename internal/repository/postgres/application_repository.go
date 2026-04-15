package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

const createApplicationQuery = `
INSERT INTO applications (manager_id, customer_name, contact_info, description, form_answer_id)
VALUES (
    (
        SELECT s.id
        FROM staff s
        LEFT JOIN applications a ON a.manager_id = s.id
            AND a.status != 'cancelled'
        LEFT JOIN bookings b ON b.manager_id = s.id
            AND b.status != 'cancelled'
        WHERE s.role IN ('manager_1', 'manager_2', 'manager_3')
        GROUP BY s.id
        ORDER BY COUNT(a.id) + COUNT(b.id) ASC
        LIMIT 1
    ),
    $1, $2, $3, $4
)
	`

const getApplicationQuery = `
	SELECT a.id, a.status, a.form_answer_id, a.customer_name, a.contact_info,
		a.description, a.created_at, a.updated_at, 
		COALESCE(a.manager_id, 0) AS manager_id,
		COALESCE(s.first_name || ' ' || s.last_name, '') AS manager_name
	FROM applications a
	LEFT JOIN staff s ON a.manager_id = s.id
	WHERE a.id = $1
`
const updateApplicationsStatus = `
UPDATE applications 
SET status = $1, updated_at = NOW()
WHERE id = $2
		`
const listApplicationsBaseQuery = `
		SELECT 
				a.id, a.status, a.customer_name, a.contact_info, a.created_at,
				COALESCE(a.manager_id, 0) AS manager_id,
				COALESCE(s.first_name || ' ' || s.last_name, '') AS manager_name,
				COUNT(*) OVER() AS total
		FROM applications a
		LEFT JOIN staff s ON s.id = a.manager_id
		`
const deleteApplication = `DELETE FROM applications WHERE id=$1`

type ApplicationRepo struct {
	db *sqlx.DB
}

func NewApplicationRepository(db *sqlx.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) CreateApplication(ctx context.Context, req *models.Application) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, createApplicationQuery,
		req.CustomerName,
		req.ContactInfo,
		req.Description,
		req.FormAnswerId)
	if err != nil {
		return fmt.Errorf("create application tx: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit application tx: %w", err)
	}

	return nil
}

func (r *ApplicationRepo) GetApplicationByID(ctx context.Context, id int64) (*models.Application, error) {
	var app dto.ApplicationDB
	err := sqlx.GetContext(ctx, r.getDB(ctx), &app, getApplicationQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrApplicationNotFound
		}
		return nil, err
	}
	return toDomainModel(&app), nil
}

func (r *ApplicationRepo) UpdateApplicationStatus(ctx context.Context, id int64, status string) error {
	result, err := r.getDB(ctx).ExecContext(ctx, updateApplicationsStatus, status, id)
	if err != nil {
		return fmt.Errorf("update application status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return models.ErrApplicationNotFound
	}

	return nil
}

func (r *ApplicationRepo) ApplicationsList(ctx context.Context, filter *models.ApplicationFilter) (*models.ApplicationList, error) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 5)
	i := 1

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("a.status = $%d", i))
		args = append(args, filter.Status)
		i++
	}
	if filter.ManagerID != 0 {
		conditions = append(conditions, fmt.Sprintf("a.manager_id = $%d", i))
		args = append(args, filter.ManagerID)
		i++
	}
	if filter.CustomerName != "" {
		conditions = append(conditions, fmt.Sprintf("a.customer_name ILIKE $%d", i))
		args = append(args, "%"+filter.CustomerName+"%")
		i++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, filter.Limit, filter.Offset)
	dataSQL := listApplicationsBaseQuery + where +
		fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", i, i+1)

	var rows []dto.ApplicationRow
	if err := r.db.SelectContext(ctx, &rows, dataSQL, args...); err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}

	apps := make([]models.Application, len(rows))
	for i, row := range rows {
		apps[i] = models.Application{
			ID:           row.ID,
			Status:       row.Status,
			ManagerID:    row.ManagerID,
			ManagerName:  row.ManagerName,
			CustomerName: row.CustomerName,
			ContactInfo:  row.ContactInfo,
			CreatedAt:    row.CreatedAt,
		}
	}

	total := 0
	if len(rows) > 0 {
		total = rows[0].Total
	}

	return &models.ApplicationList{
		Items:  apps,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (r *ApplicationRepo) DeleteApplication(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, deleteApplication, id)
	if err != nil {
		return fmt.Errorf("delete application: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return models.ErrApplicationNotFound
	}

	return nil
}

func toDomainModel(a *dto.ApplicationDB) *models.Application {
	return &models.Application{
		ID:           a.ID,
		Status:       a.Status,
		ManagerID:    a.ManagerID,
		ManagerName:  a.ManagerName,
		FormAnswerId: a.FormAnswerId,
		CustomerName: a.CustomerName,
		ContactInfo:  a.ContactInfo,
		Description:  a.Description,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func (r *ApplicationRepo) getDB(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctxutil.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
