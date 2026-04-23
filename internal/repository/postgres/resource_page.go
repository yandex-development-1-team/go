package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
)

type ResourcePageRepository struct {
	db *sqlx.DB
}

func NewResourcePageRepo(db *sqlx.DB) *ResourcePageRepository {
	return &ResourcePageRepository{db: db}
}

func (r *ResourcePageRepository) GetAll(ctx context.Context) ([]models.ResourcePage, error) {
	var rows []dto.ResourcePageDB
	err := r.db.SelectContext(ctx, &rows, `
			SELECT slug, title, content, links, updated_at
			FROM resource_pages
	`)
	if err != nil {
		return nil, fmt.Errorf("get all resource pages: %w", err)
	}

	pages := make([]models.ResourcePage, 0, len(rows))
	for _, row := range rows {
		p, err := toDomain(&row)
		if err != nil {
			return nil, fmt.Errorf("map resource page %s: %w", row.Slug, err)
		}
		pages = append(pages, *p)
	}
	return pages, nil
}

func (r *ResourcePageRepository) GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error) {
	var row dto.ResourcePageDB
	err := r.db.GetContext(ctx, &row, `
			SELECT slug, title, content, links, updated_at
			FROM resource_pages
			WHERE slug = $1
	`, slug)
	if err != nil {
		return nil, fmt.Errorf("get resource page by slug: %w", err)
	}
	return toDomain(&row)
}

func (r *ResourcePageRepository) Update(ctx context.Context, slug string, page models.ResourcePage) (*models.ResourcePage, error) {
	for i := range page.Links {
		if page.Links[i].ID == "" {
			page.Links[i].ID = uuid.New().String()
		}
	}

	linksJSON, err := json.Marshal(page.Links)
	if err != nil {
		return nil, fmt.Errorf("marshal links: %w", err)
	}

	var row dto.ResourcePageDB
	err = sqlx.GetContext(ctx, r.getDB(ctx), &row, `
			UPDATE resource_pages
			SET title = $1, content = $2, links = $3, updated_at = NOW()
			WHERE slug = $4
			RETURNING slug, title, content, links, updated_at
	`, page.Title, page.Content, linksJSON, slug)
	if err != nil {
		return nil, fmt.Errorf("update resource page: %w", err)
	}

	result, err := toDomain(&row)
	return result, err
}

func (r *ResourcePageRepository) Clear(ctx context.Context, slug string) (*models.ResourcePage, error) {
	var row dto.ResourcePageDB
	err := r.db.GetContext(ctx, &row, `
			UPDATE resource_pages
			SET content = '', links = '[]'::jsonb, updated_at = NOW()
			WHERE slug = $1
			RETURNING slug, title, content, links, updated_at
	`, slug)
	if err != nil {
		return nil, fmt.Errorf("clear resource page: %w", err)
	}
	return toDomain(&row)
}

func (r *ResourcePageRepository) DeleteLink(ctx context.Context, slug string, id string) (*models.ResourcePage, error) {
	var row dto.ResourcePageDB
	err := r.db.GetContext(ctx, &row, `
			UPDATE resource_pages
			SET links = COALESCE(
				(
					SELECT jsonb_agg(el)
					FROM jsonb_array_elements(links) el
					WHERE el->>'id' != $1
				),
				'[]'::jsonb
			),
			updated_at = NOW()
			WHERE slug = $2
			RETURNING slug, title, content, links, updated_at
	`, id, slug)
	if err != nil {
		return nil, fmt.Errorf("delete link: %w", err)
	}
	return toDomain(&row)
}

func toDomain(row *dto.ResourcePageDB) (*models.ResourcePage, error) {
	var links []models.ResourcePageLink
	if len(row.Links) > 0 {
		if err := json.Unmarshal(row.Links, &links); err != nil {
			return nil, fmt.Errorf("unmarshal links: %w", err)
		}
	}
	if links == nil {
		links = []models.ResourcePageLink{}
	}

	return &models.ResourcePage{
		Slug:      row.Slug,
		Title:     row.Title,
		Content:   row.Content,
		Links:     links,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *ResourcePageRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx, ok := ctxutil.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}
