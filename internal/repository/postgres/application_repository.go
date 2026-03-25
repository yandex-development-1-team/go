package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const insertApplicationQuery = `
	INSERT INTO applications (type, source, customer_name, contact_info, project_name, box_id, special_project_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, type, source, status, customer_name, contact_info,
	          project_name, box_id, special_project_id, manager_id, created_at, updated_at`

type ApplicationRepo struct {
	db *sqlx.DB
}

func NewApplicationRepository(db *sqlx.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error) {
	const operation = "create_application"

	if req == nil || req.CustomerName == "" || req.ContactInfo == "" || !req.Type.Valid() || !req.Source.Valid() {
		return nil, models.ErrInvalidInput
	}

	var app models.Application
	err := r.db.QueryRowxContext(ctx, insertApplicationQuery,
		req.Type, req.Source, req.CustomerName, req.ContactInfo,
		req.ProjectName, req.BoxID, req.SpecialProjectID,
	).StructScan(&app)
	if err != nil {
		return nil, repository.CheckDBError(operation, err)
	}
	return &app, nil
}

func (r *ApplicationRepo) GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
	const operation = "get_applications"

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	where, args := buildApplicationWhere(filter)

	var total int
	if err := r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM applications %s`, where), args...).Scan(&total); err != nil {
		return nil, 0, repository.CheckDBError(operation, err)
	}

	listQuery := fmt.Sprintf(`
		SELECT id, type, source, status, customer_name, contact_info,
		       project_name, box_id, special_project_id, manager_id, created_at, updated_at
		FROM applications %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)

	var apps []models.Application
	if err := r.db.SelectContext(ctx, &apps, listQuery, append(args, filter.Limit, filter.Offset)...); err != nil {
		return nil, 0, repository.CheckDBError(operation, err)
	}
	return apps, total, nil
}

const selectApplicationByIDQuery = `
	SELECT id, type, source, status, customer_name, contact_info,
	       project_name, box_id, special_project_id, manager_id, created_at, updated_at
	FROM applications
	WHERE id = $1`

func (r *ApplicationRepo) GetApplicationByID(ctx context.Context, id int64) (*models.Application, error) {
	const operation = "get_application_by_id"

	var app models.Application
	err := r.db.QueryRowxContext(ctx, selectApplicationByIDQuery, id).StructScan(&app)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrApplicationNotFound
		}
		return nil, repository.CheckDBError(operation, err)
	}
	return &app, nil
}

func (r *ApplicationRepo) UpdateApplication(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error) {
	const operation = "update_application"

	if req == nil || !req.HasUpdates() {
		return nil, models.ErrInvalidInput
	}

	var setClauses []string
	var args []interface{}

	addSet := func(col string, val interface{}) {
		args = append(args, val)
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, len(args)))
	}

	if req.Status != nil {
		addSet("status", *req.Status)
	}
	if req.ContactInfo != nil {
		addSet("contact_info", *req.ContactInfo)
	}
	if req.BoxID != nil {
		addSet("box_id", *req.BoxID)
	}
	if req.SpecialProjectID != nil {
		addSet("special_project_id", *req.SpecialProjectID)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE applications
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, type, source, status, customer_name, contact_info,
		          project_name, box_id, special_project_id, manager_id, created_at, updated_at`,
		strings.Join(setClauses, ", "), len(args))

	var app models.Application
	err := r.db.QueryRowxContext(ctx, query, args...).StructScan(&app)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrApplicationNotFound
		}
		return nil, repository.CheckDBError(operation, err)
	}
	return &app, nil
}

func (r *ApplicationRepo) DeleteApplication(ctx context.Context, id int64) error {
	const operation = "delete_application"

	result, err := r.db.ExecContext(ctx, `DELETE FROM applications WHERE id = $1`, id)
	if err != nil {
		return repository.CheckDBError(operation, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return repository.CheckDBError(operation, err)
	}
	if affected == 0 {
		return models.ErrApplicationNotFound
	}
	return nil
}

func buildApplicationWhere(f models.ApplicationFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	add := func(clause string, val interface{}) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf(clause, len(args)))
	}

	if f.Status != nil {
		add("status = $%d", *f.Status)
	}
	if f.Type != nil {
		add("type = $%d", *f.Type)
	}
	if f.ManagerID != nil {
		add("manager_id = $%d", *f.ManagerID)
	}
	if f.DateFrom != nil {
		add("created_at >= $%d", *f.DateFrom)
	}
	if f.DateTo != nil {
		add("created_at <= $%d", *f.DateTo)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}
