package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

type specialProjectRepo struct {
	db *sqlx.DB
}

const (
	createSpecProjectQuery = `
    WITH new_image AS (
    UPDATE files
    SET is_active = true
    WHERE url = $3        -- было $4, должно быть $3 (image)
        AND $3 IS NOT NULL
        AND $3 != ''
)
	INSERT INTO special_projects (title, description, image, status)
	VALUES ($1, $2, $3, $4::spec_type)
	RETURNING id, title, description, image, status, created_at, updated_at`

	updateSpecProjectQuery = `
    WITH
    old_image AS (
        SELECT image FROM special_projects WHERE id = $5
    ),
    deactivate AS (
        UPDATE files f
        SET is_active = false
        FROM old_image o
        WHERE f.url = o.image
            AND o.image IS NOT NULL
            AND o.image != ''
            AND $3::text IS NOT NULL
			AND $3::text != ''
			AND o.image != $3::text
    ),
    activate AS (
        UPDATE files
        SET is_active = true
        WHERE url = $3::text
			AND $3::text IS NOT NULL
    		AND $3::text != ''
    )
    UPDATE special_projects
    SET
        title       = COALESCE($1, title),
        description = COALESCE($2, description),
        image = COALESCE($3, image),
        status = COALESCE($4, status),
        updated_at  = NOW()
    WHERE id = $5
    RETURNING id, title, description, image, status, created_at, updated_at`

	getSpecProjectByIDQuery = `
	SELECT id, title, description, image, status, created_at, updated_at 
	FROM special_projects 
	WHERE id = $1 AND deleted_at IS NULL`

	listSpecProjectsBaseQuery      = `SELECT id, title, description, image, status, created_at, updated_at FROM special_projects WHERE deleted_at IS NULL`
	listSpecProjectsCountBaseQuery = `SELECT COUNT(1) FROM special_projects WHERE deleted_at IS NULL`

	deleteSpecProjectQuery = `
    WITH deactivate_image AS (
        UPDATE files f
        SET is_active = false
        FROM special_projects sp
        WHERE sp.id = $1
            AND f.url = sp.image
            AND sp.image IS NOT NULL
            AND sp.image != ''
    )
    UPDATE special_projects 
    SET deleted_at = now(), updated_at = now(), status = 'inactive'
    WHERE id = $1 AND deleted_at IS NULL`
)

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error) {
	var result models.SpecialProjectDB

	err := r.db.QueryRowContext(ctx, createSpecProjectQuery,
		proj.Title,
		proj.Description,
		proj.Image,
		proj.Status,
	).Scan(
		&result.ID,
		&result.Title,
		&result.Description,
		&result.Image,
		&result.Status,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w", models.ErrSpecialProjectAlreadyExists)
		}
		return nil, fmt.Errorf("repo create: %w", err)
	}
	return &result, nil
}

func (r *specialProjectRepo) GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
	var proj models.SpecialProjectDB

	err := r.db.GetContext(ctx, &proj, getSpecProjectByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	return &proj, nil
}

func (r *specialProjectRepo) List(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error) {
	// Base query selecting only required fields for the list endpoint
	baseQuery := listSpecProjectsBaseQuery
	baseCountQuery := listSpecProjectsCountBaseQuery

	args := make(map[string]interface{})
	countArgs := make(map[string]interface{})

	// Apply status filter if provided
	if statusFilter != "" {
		baseQuery += " AND status = :status"
		baseCountQuery += " AND status = :status"
		args["status"] = statusFilter
		countArgs["status"] = statusFilter
	}

	// Apply full-text search if query is provided
	if searchQuery != "" {
		// Use PostgreSQL's built-in full-text search with to_tsvector
		// This requires the GIN index created earlier
		searchFragment := ` AND to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('russian', :search)`
		baseQuery += searchFragment
		baseCountQuery += searchFragment
		args["search"] = searchQuery
		countArgs["search"] = searchQuery
	}

	countStmt, err := r.db.PrepareNamedContext(ctx, baseCountQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare count query: %w", err)
	}
	defer func() { _ = countStmt.Close() }()

	var total int
	err = countStmt.GetContext(ctx, &total, countArgs)
	if err != nil {
		return nil, 0, fmt.Errorf("repo count: %w", err)
	}

	// Order results by creation date descending
	baseQuery += " ORDER BY updated_at DESC LIMIT :limit OFFSET :offset"
	args["limit"] = limit
	args["offset"] = offset
	// Prepare the named query
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare named query: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	var projects []*models.SpecialProjectDB
	err = stmt.SelectContext(ctx, &projects, args)
	if err != nil {
		return nil, 0, fmt.Errorf("repo list: %w", err)
	}

	return projects, total, nil
}

func (r *specialProjectRepo) Update(ctx context.Context, id int64, specialProject *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
	var updatedSpecialProject models.SpecialProjectDB

	err := r.db.QueryRowContext(ctx, updateSpecProjectQuery,
		specialProject.Title,
		specialProject.Description,
		specialProject.Image,
		specialProject.Status,
		id,
	).Scan(
		&updatedSpecialProject.ID,
		&updatedSpecialProject.Title,
		&updatedSpecialProject.Description,
		&updatedSpecialProject.Image,
		&updatedSpecialProject.Status,
		&updatedSpecialProject.CreatedAt,
		&updatedSpecialProject.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, fmt.Errorf("failed to update special project: %w", err)
	}
	return &updatedSpecialProject, nil
}

func (r specialProjectRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, deleteSpecProjectQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete special project: %w", err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return models.ErrSpecialProjectNotFound
	}

	return nil
}
