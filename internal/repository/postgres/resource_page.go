package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/repository/models"
)

type ResourcePageRepo struct {
	db *sqlx.DB
}

func NewResourcePageRepo(db *sqlx.DB) *ResourcePageRepo {
	return &ResourcePageRepo{db: db}
}

// GetBySlug
func (r *ResourcePageRepo) GetBySlug(ctx context.Context, slug string) (*models.ResourcePageDB, error) {
	var dbPage models.ResourcePageDB

	err := r.db.QueryRowxContext(ctx, "SELECT slug, title, content, links, updated_at FROM resource_pages WHERE slug = $1", slug).StructScan(&dbPage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resource page with slug '%s' not found", slug)
		}
	}
	return &dbPage, nil
}

// UpdatePage обновляет существующую страницу по slug.
func (r *ResourcePageRepo) UpdatePage(ctx context.Context, page *models.ResourcePageDB) error {

	query := `UPDATE resource_pages
		SET
			title = $2,
			content = $3,
			links = $4,
			updated_at = NOW()
		WHERE slug = $1;
	`

	res, err := r.db.ExecContext(ctx, query, page.Slug, page.Title, page.Content, page.LinksJSON)
	if err != nil {
		return fmt.Errorf("error updating resource page: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during update resource page with slug '%s': %v", page.Slug, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("resource page with slug '%s' not found", page.Slug)
	}

	return nil
}
