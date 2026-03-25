package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type ResourcePageRepository struct {
	db *sqlx.DB
}

func NewResourcePageRepo(db *sqlx.DB) *ResourcePageRepository {
	return &ResourcePageRepository{db: db}
}

func (r *ResourcePageRepository) GetBySlug(ctx context.Context, slug string) (*models.ResourcePage, error) {
	return r.GetBySlugTx(ctx, r.db, slug, false)
}

func (r *ResourcePageRepository) GetBySlugTx(ctx context.Context, queryable repository.Queryable, slug string, lockForUpdate bool) (*models.ResourcePage, error) {
	var dbModel models.ResourcePageDB
	query := "SELECT slug, title, content, links, updated_at FROM resource_pages WHERE slug = $1"
	if lockForUpdate {
		query += " FOR UPDATE"
	}
	query += ";"

	err := queryable.GetContext(ctx, &dbModel, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrResourcePageNotFound
		}
		return nil, err
	}

	domainModel, err := toDomainDB(&dbModel)
	if err != nil {
		return nil, fmt.Errorf("error converting DB model to domain model: %w", err)
	}

	return domainModel, nil
}

func (r *ResourcePageRepository) UpdatePageContentAndLinksTx(ctx context.Context, tx *sqlx.Tx, slug string, title string, content string, links []models.ResourcePageLink) error {
	domainModel := &models.ResourcePage{
		Slug:    slug,
		Title:   title,
		Content: content,
		Links:   links,
	}
	dbModel, err := toRepoDB(domainModel)
	if err != nil {
		return fmt.Errorf("error converting domain model to DB model for update: %w", err)
	}

	query := `
		UPDATE resource_pages
		SET
			title = $2,
			content = $3,
			links = $4,
			updated_at = NOW()
		WHERE slug = $1;
	`

	result, err := tx.ExecContext(ctx, query, dbModel.Slug, dbModel.Title, dbModel.Content, dbModel.LinksJSON)
	if err != nil {
		return fmt.Errorf("error updating resource page in transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error updating resource page in transaction: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrResourcePageNotFound
	}

	return nil
}

func (r *ResourcePageRepository) GetAllSummaries(ctx context.Context) ([]*models.ResourcePage, error) {
	query := "SELECT slug, title, updated_at FROM resource_pages ORDER BY updated_at DESC;"
	var dbModels []struct {
		Slug      string    `db:"slug"`
		Title     string    `db:"title"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	err := r.db.SelectContext(ctx, &dbModels, query)
	if err != nil {
		return nil, fmt.Errorf("error querying resource page summaries: %w", err)
	}

	domainModels := make([]*models.ResourcePage, len(dbModels))
	for i, dbMod := range dbModels {
		domainModels[i] = &models.ResourcePage{
			Slug:      dbMod.Slug,
			Title:     dbMod.Title,
			UpdatedAt: dbMod.UpdatedAt.Format(time.RFC3339),
		}
	}

	return domainModels, nil
}

func (r *ResourcePageRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func toDomainDB(dbModel *models.ResourcePageDB) (*models.ResourcePage, error) {
	if dbModel == nil {
		return nil, nil
	}

	var links []models.ResourcePageLink
	if len(dbModel.LinksJSON) > 0 {
		if err := json.Unmarshal(dbModel.LinksJSON, &links); err != nil {
			return nil, err
		}
	}

	var content string
	if dbModel.Content != nil {
		content = *dbModel.Content
	}

	return &models.ResourcePage{
		Slug:      dbModel.Slug,
		Title:     dbModel.Title,
		Content:   content,
		Links:     links,
		LinksJSON: dbModel.LinksJSON,
		UpdatedAt: dbModel.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func toRepoDB(domainModel *models.ResourcePage) (*models.ResourcePageDB, error) {
	if domainModel == nil {
		return nil, nil
	}

	linksJSON, err := json.Marshal(domainModel.Links)
	if err != nil {
		return nil, err
	}

	var contentPtr *string
	if domainModel.Content != "" {
		contentPtr = &domainModel.Content
	}

	return &models.ResourcePageDB{
		Slug:      domainModel.Slug,
		Title:     domainModel.Title,
		Content:   contentPtr,
		LinksJSON: linksJSON,
	}, nil
}
