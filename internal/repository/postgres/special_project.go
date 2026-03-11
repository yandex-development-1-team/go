package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/specialproject"
)

const (
	insertSpecialProjectQuery = `
		INSERT INTO special_projects (title, description, image, is_active_in_bot)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, image, is_active_in_bot, created_at, updated_at`

	getSpecialProjectByIDQuery = `
		SELECT id, title, description, image, is_active_in_bot, created_at, updated_at
		FROM special_projects
		WHERE id = $1`

	listSpecialProjectsBaseQuery = `
		SELECT id, title, is_active_in_bot
		FROM special_projects`

	updateSpecialProjectQuery = `
		UPDATE special_projects
		SET title = $1, description = $2, image = $3, is_active_in_bot = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, title, description, image, is_active_in_bot, created_at, updated_at`

	deactivateApplicationsQuery = `
		UPDATE applications
		SET status = 'cancelled', updated_at = NOW()
		WHERE special_project_id = $1 AND status IN ('queue', 'in_progress')`

	deleteSpecialProjectQuery = `
		DELETE FROM special_projects
		WHERE id = $1`
)

type specialProjectRepo struct {
	db *sqlx.DB
}

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Error("failed to rollback transaction", zap.Error(rollbackErr))
		}
	}()

	var out specialproject.DB
	err = tx.QueryRowxContext(ctx, insertSpecialProjectQuery,
		proj.Title, proj.Description, proj.Image, proj.IsActiveInBot,
	).StructScan(&out)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return nil, specialproject.ErrAlreadyExists
		}
		return nil, fmt.Errorf("repo create: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return &out, nil
}

func (r *specialProjectRepo) GetByID(ctx context.Context, id int64) (*specialproject.DB, error) {
	var proj specialproject.DB
	err := r.db.GetContext(ctx, &proj, getSpecialProjectByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, specialproject.ErrNotFound
		}
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	return &proj, nil
}

func (r *specialProjectRepo) List(ctx context.Context, statusFilter *bool, searchQuery string) ([]*specialproject.DB, error) {
	baseQuery := listSpecialProjectsBaseQuery
	var args []interface{}

	if statusFilter != nil {
		baseQuery += fmt.Sprintf(" AND is_active_in_bot = $%d", len(args)+1)
		args = append(args, *statusFilter)
	}
	if searchQuery != "" {
		baseQuery += fmt.Sprintf(" AND to_tsvector('russian', coalesce(title, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('russian', $%d)", len(args)+1)
		args = append(args, searchQuery)
	}
	baseQuery += " ORDER BY updated_at DESC"

	var projects []*specialproject.DB
	err := r.db.SelectContext(ctx, &projects, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("repo list: %w", err)
	}
	return projects, nil
}

func (r *specialProjectRepo) UpdateSpecialProject(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Error("failed to rollback transaction", zap.Error(rollbackErr))
		}
	}()

	var out specialproject.DB
	err = tx.QueryRowxContext(ctx, updateSpecialProjectQuery,
		update.Title, update.Description, update.Image, update.IsActiveInBot, id,
	).StructScan(&out)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, specialproject.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update special project: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return &out, nil
}

func (r *specialProjectRepo) DeleteSpecialProject(ctx context.Context, id int64) error {
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
		return fmt.Errorf("failed to deactivate applications: %w", err)
	}
	res, err := tx.ExecContext(ctx, deleteSpecialProjectQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete special project: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return specialproject.ErrNotFound
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
