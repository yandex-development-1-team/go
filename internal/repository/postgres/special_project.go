package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/specialproject"
	"go.uber.org/zap"
)

type specialProjectRepo struct {
	db *sqlx.DB
}

const (
	createQuery = `
		INSERT INTO special_projects (title, description, image, is_active_in_bot)
		VALUES (:title, :description, :image, :is_active_in_bot)
		RETURNING id, title, description, image, is_active_in_bot, created_at, updated_at
	`
	updateQuery = `
		UPDATE special_projects 
		SET title = $1, description = $2, image = $3, is_active_in_bot = $4
		WHERE id = $5
		RETURNING title, description, image, is_active_in_bot`

	getByIDQuery = `SELECT * FROM special_projects WHERE id = $1`

	listBaseQuery      = `SELECT id, title, is_active_in_bot FROM special_projects WHERE 1=1`
	listCountBaseQuery = `SELECT COUNT(1) FROM special_projects WHERE 1=1`

	deactivateApplicationsQuery = `
        UPDATE applications 
		SET status = 'cancelled', updated_at = now()
		WHERE special_project_id = $1 AND status IN ('queue', 'in_progress')`

	deleteQuery = `
    UPDATE special_projects 
    SET deleted_at = now(), updated_at = now(), is_active_in_bot = false
    WHERE id = $1 AND deleted_at IS NULL`
)

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error) {

	err := r.db.QueryRowxContext(ctx, createQuery, proj).StructScan(proj)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // duplicate key
			return nil, fmt.Errorf("special project already exists: %w", err)
		}
		return nil, fmt.Errorf("repo create: %w", err)
	}
	return proj, nil

}

func (r *specialProjectRepo) GetByID(ctx context.Context, id int64) (*specialproject.DB, error) {
	var proj specialproject.DB

	err := r.db.GetContext(ctx, &proj, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, specialproject.ErrNotFound
		}
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	return &proj, nil
}

func (r *specialProjectRepo) List(ctx context.Context, statusFilter *bool, searchQuery string, limit, offset int) ([]*specialproject.DB, int, error) {
	// Base query selecting only required fields for the list endpoint
	baseQuery := listBaseQuery
	baseCountQuery := listCountBaseQuery

	args := make(map[string]interface{})
	countArgs := make(map[string]interface{})

	// Apply status filter if provided
	if statusFilter != nil {
		baseQuery += " AND is_active_in_bot = :status"
		baseCountQuery += " AND is_active_in_bot = :status"
		args["status"] = *statusFilter
		countArgs["status"] = *statusFilter
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
	defer countStmt.Close()

	var total int
	err = countStmt.GetContext(ctx, &total, countArgs)
	if err != nil {
		return nil, 0, fmt.Errorf("repo count: %w", err)
	}

	// Order results by creation date descending
	baseQuery += " ORDER BY updated_at DESC LIMIT :limit OFFSET :offset"

	// Prepare the named query
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to prepare named query: %w", err)
	}
	defer stmt.Close()

	var projects []*specialproject.DB
	err = stmt.SelectContext(ctx, &projects, args)
	if err != nil {
		return nil, 0, fmt.Errorf("repo list: %w", err)
	}

	return projects, total, nil
}

func (r *specialProjectRepo) Update(ctx context.Context, id int64, specialProject *specialproject.Update) (*specialproject.DB, error) {

	var updatedSpecialProject specialproject.DB

	err := r.db.QueryRowContext(ctx, updateQuery,
		specialProject.Title,
		specialProject.Description,
		specialProject.Image,
		specialProject.IsActiveInBot,
		id,
	).Scan(&updatedSpecialProject.Title, &updatedSpecialProject.Description, &updatedSpecialProject.Image, &updatedSpecialProject.IsActiveInBot)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, specialproject.ErrNotFound
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

	res, err := tx.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete special project: %w", err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return specialproject.ErrNotFound
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
