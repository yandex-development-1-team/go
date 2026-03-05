package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
)

const (
	updateSpecialProjectQuery = `
		UPDATE special_projects 
		SET title = $1, description = $2, image = $3, is_active_in_bot = $4
		WHERE id = $5
		RETURNING title, description, image, is_active_in_bot`

	deactivateApplicationsQuery = `
        UPDATE applications 
		SET status = 'cancelled', updated_at = now()
		WHERE special_project_id = $1 AND status IN ('queue', 'in_progress')`

	deleteSpecialProject = `
    UPDATE special_projects 
    SET deleted_at = now(), updated_at = now(), is_active_in_bot = false
    WHERE id = $1 AND deleted_at IS NULL`
)

type SpecialProjectRepository interface {
	UpdateSpecialProject(ctx context.Context, id int, specialProject models.SpecialProject) error
	DeleteSpecialProject(ctx context.Context, id int) error
}

type SpecialProjectRepo struct {
	db *sqlx.DB
}

func (sp *SpecialProjectRepo) UpdateSpecialProject(ctx context.Context, id int, specialProject models.SpecialProject) (models.SpecialProject, error) {
	var updatedSpecialProject models.SpecialProject
	err := sp.db.QueryRowContext(ctx, updateSpecialProjectQuery,
		specialProject.Title,
		specialProject.Description,
		specialProject.Image,
		specialProject.Is_active_in_bot,
		id,
	).Scan(&updatedSpecialProject.Title, &updatedSpecialProject.Description, &updatedSpecialProject.Image, &updatedSpecialProject.Is_active_in_bot)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SpecialProject{}, models.ErrSpecialProjectNotFound
		}
		return models.SpecialProject{}, fmt.Errorf("failed to update special project: %w", err)
	}
	return updatedSpecialProject, nil
}

func (sp *SpecialProjectRepo) DeleteSpecialProject(ctx context.Context, id int) error {
	tx, err := sp.db.BeginTxx(ctx, nil)
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

	res, err := tx.ExecContext(ctx, deleteSpecialProject, id)
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
