package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	createApplicationQuery = `
		INSERT INTO applications (type, source, customer_name, contact_info, project_name, box_id, special_project_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, type, source, status, customer_name, contact_info,
		          project_name, box_id, special_project_id, manager_id, created_at, updated_at`

	countApplicationsBaseQuery = `SELECT COUNT(*) FROM applications`
	listApplicationsColumns    = `id, type, source, status, customer_name, contact_info,
		project_name, box_id, special_project_id, manager_id, created_at, updated_at`
)

type ApplicationRepository interface {
	CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error)
	GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error)
}

type ApplicationRepo struct {
	db *sqlx.DB
}

var _ ApplicationRepository = (*ApplicationRepo)(nil)

func NewApplicationRepository(db *sqlx.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error) {
	const operation = "create_application"

	if req == nil || req.CustomerName == "" || req.ContactInfo == "" || !req.Type.Valid() || !req.Source.Valid() {
		return nil, models.ErrInvalidInput
	}

	var app models.Application
	err := r.db.QueryRowxContext(ctx, createApplicationQuery,
		req.Type, req.Source, req.CustomerName, req.ContactInfo,
		req.ProjectName, req.BoxID, req.SpecialProjectID,
	).StructScan(&app)
	if err != nil {
		return nil, checkError(operation, err)
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
	if err := r.db.QueryRowContext(ctx, countApplicationsBaseQuery+" "+where, args...).Scan(&total); err != nil {
		return nil, 0, checkError(operation, err)
	}

	listQuery := fmt.Sprintf("SELECT %s FROM applications %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		listApplicationsColumns, where, len(args)+1, len(args)+2)

	var apps []models.Application
	if err := r.db.SelectContext(ctx, &apps, listQuery, append(args, filter.Limit, filter.Offset)...); err != nil {
		return nil, 0, checkError(operation, err)
	}
	return apps, total, nil
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
