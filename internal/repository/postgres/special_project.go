package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	insertSpecialProjectQuery = `
		INSERT INTO special_projects (title, description, image, is_active_in_bot)
		VALUES (:title, :description, :image, :is_active_in_bot)
		RETURNING id, title, description, image, is_active_in_bot, created_at, updated_at`

	getSpecialProjectByIDQuery = `SELECT * FROM special_projects WHERE id = $1`

	listSpecialProjectsBaseQuery = `SELECT id, title, is_active_in_bot FROM special_projects WHERE 1=1`
)

type specialProjectRepo struct {
	db *sqlx.DB
}

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
	err := r.db.QueryRowxContext(ctx, insertSpecialProjectQuery, proj).StructScan(proj)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // duplicate key
			return nil, fmt.Errorf("special project already exists: %w", err)
		}
		return nil, fmt.Errorf("repo create: %w", err)
	}
	return proj, nil

}

func (r *specialProjectRepo) GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
	var proj models.SpecialProjectDB
	err := r.db.GetContext(ctx, &proj, getSpecialProjectByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSpecProjNotFound
		}
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	return &proj, nil
}

func (r *specialProjectRepo) List(ctx context.Context, statusFilter *bool, searchQuery string) ([]*models.SpecialProjectDB, error) {
	baseQuery := listSpecialProjectsBaseQuery
	args := make(map[string]interface{})

	// Apply status filter if provided
	if statusFilter != nil {
		baseQuery += " AND is_active_in_bot = :status"
		args["status"] = *statusFilter
	}

	// Apply full-text search if query is provided
	if searchQuery != "" {
		// Use PostgreSQL's built-in full-text search with to_tsvector
		// This requires the GIN index created earlier
		baseQuery += ` AND to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('russian', :search)`
		args["search"] = searchQuery
	}

	// Order results by creation date descending
	baseQuery += " ORDER BY updated_at DESC"

	// Prepare the named query
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare named query: %w", err)
	}
	defer stmt.Close()

	var projects []*models.SpecialProjectDB
	err = stmt.SelectContext(ctx, &projects, args)
	if err != nil {
		return nil, fmt.Errorf("repo list: %w", err)
	}

	return projects, nil
}
