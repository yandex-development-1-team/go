package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
)

type specialProjectRepo struct {
	db *sqlx.DB
}

const (
	createSpecProjectQuery = `
		INSERT INTO special_projects (title, description, image, status)
		VALUES (:title, :description, :image, :status)
		RETURNING id, title, description, image, status, created_at, updated_at
	`
	updateSpecProjectQuery = `
		UPDATE special_projects 
		SET title = $1, description = $2, image = $3, status = $4
		WHERE id = $5
		RETURNING title, description, image, status`

	getSpecProjectByIDQuery = `SELECT * FROM special_projects WHERE id = $1`

	listSpecProjectsBaseQuery      = `SELECT id, title, status FROM special_projects WHERE 1=1`
	listSpecProjectsCountBaseQuery = `SELECT COUNT(1) FROM special_projects WHERE 1=1`

	deactivateApplicationsQuery = `
        UPDATE applications 
		SET status = 'done', updated_at = now()
		WHERE special_project_id = $1 AND status IN ('queue', 'in_progress')`

	deleteSpecProjectQuery = `
    UPDATE special_projects 
    SET deleted_at = now(), updated_at = now(), status = 'inactive'
    WHERE id = $1 AND deleted_at IS NULL`
)

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
	rows, err := r.db.NamedQueryContext(ctx, createSpecProjectQuery, proj)
	if err == nil {
		if rows.Next() {
			err = rows.StructScan(proj)
		}
		_ = rows.Close()
	}
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("%w", models.ErrSpecialProjectAlreadyExists)
		}
		return nil, fmt.Errorf("repo create: %w", err)
	}
	return proj, nil
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
	).Scan(&updatedSpecialProject.Title, &updatedSpecialProject.Description, &updatedSpecialProject.Image, &updatedSpecialProject.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSpecialProjectNotFound
		}
		return nil, fmt.Errorf("failed to update special project: %w", err)
	}
	return &updatedSpecialProject, nil
}

func (r specialProjectRepo) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Error("failed to rollback transaction", zap.Error(rollbackErr))
		}
	}()

	if _, err = tx.ExecContext(ctx, deactivateApplicationsQuery, id); err != nil {
		return fmt.Errorf("failed to deactivate special project in applications: %w", err)
	}

	res, err := tx.ExecContext(ctx, deleteSpecProjectQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete special project: %w", err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return models.ErrSpecialProjectNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
