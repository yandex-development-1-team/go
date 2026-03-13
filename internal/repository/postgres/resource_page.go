package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/resoursepage"
)

type ResourcePageRepository struct {
	db *sqlx.DB
}

func NewResourcePageRepo(db *sqlx.DB) *ResourcePageRepository {
	return &ResourcePageRepository{db: db}
}

// GetBySlug
func (r *ResourcePageRepository) GetBySlug(ctx context.Context, slug string) (*resoursepage.DB, error) {
	var dbPage resoursepage.DB

	err := r.db.QueryRowxContext(ctx, "SELECT slug, title, content, links, updated_at FROM resource_pages WHERE slug = $1", slug).StructScan(&dbPage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resource page with slug '%s' not found", slug)
		}
	}
	return &dbPage, nil
}

// UpdatePage обновляет существующую страницу по slug.
func (r *ResourcePageRepository) Update(ctx context.Context, page *resoursepage.DB) error {

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
