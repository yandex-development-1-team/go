package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yandex-development-1-team/go/internal/repository/models"
)

type specialProjectRepo struct {
	db *sqlx.DB
}

func NewSpecialProjectRepository(db *sqlx.DB) *specialProjectRepo {
	return &specialProjectRepo{db: db}
}

func (r *specialProjectRepo) Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
	query := `
		INSERT INTO special_projects (title, description, image, is_active_in_bot)
		VALUES (:title, :description, :image, :is_active_in_bot)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowxContext(ctx, query, proj).StructScan(proj)
	if err != nil {
		return nil, fmt.Errorf("repo create: %w", err)
	}
	return proj, nil
}

func (r *specialProjectRepo) GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
	query := `SELECT * FROM special_projects WHERE id = $1`
	var proj models.SpecialProjectDB

	err := r.db.GetContext(ctx, &proj, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSpecProjNotFound
		}
		return nil, fmt.Errorf("repo get by id: %w", err)
	}
	return &proj, nil
}

func (r *specialProjectRepo) List(ctx context.Context, statusFilter *bool, searchQuery string) ([]*models.SpecialProjectDB, error) {
	// Базовый запрос
	query := `SELECT id, title, is_active_in_bot FROM special_projects WHERE 1=1`
	args := make(map[string]interface{})

	// Фильтр по статусу
	if statusFilter != nil {
		query += " AND is_active_in_bot = :status"
		args["status"] = *statusFilter
	}

	// Полнотекстовый поиск (упрощенная версия через ILIKE для примера,
	// в продакшене лучше использовать to_tsvector как в индексе)
	if searchQuery != "" {
		query += " AND (title ILIKE :search OR description ILIKE :search)"
		args["search"] = "%" + searchQuery + "%"
	}

	// Добавляем ORDER BY
	query += " ORDER BY created_at DESC"

	var projects []*models.SpecialProjectDB
	// Используем NamedQuery для безопасной подстановки аргументов
	//namedQuery := sqlx.Rebind(sqlx.DOLLAR, query) // Rebind для Postgres ($1, $2)
	// Для NamedQuery нужен особый подход, если используем map.
	// Проще собрать args slice вручную для позиционных аргументов или использовать builder.
	// Ниже вариант с ручным сбором аргументов для простоты понимания без_named_query сложностей:

	finalQuery := `SELECT id, title, is_active_in_bot FROM special_projects WHERE 1=1`
	var finalArgs []interface{}
	argCount := 0

	if statusFilter != nil {
		argCount++
		finalQuery += fmt.Sprintf(" AND is_active_in_bot = $%d", argCount)
		finalArgs = append(finalArgs, *statusFilter)
	}

	if searchQuery != "" {
		argCount++
		finalQuery += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argCount, argCount)
		searchParam := "%" + searchQuery + "%"
		finalArgs = append(finalArgs, searchParam)
	}

	finalQuery += " ORDER BY created_at DESC"

	err := r.db.SelectContext(ctx, &projects, finalQuery, finalArgs...)
	if err != nil {
		return nil, fmt.Errorf("repo list: %w", err)
	}

	return projects, nil
}
