package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository"
)

type mockSpecialProjectRepo struct {
	createFn  func(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error)
	getByIDFn func(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	listFn    func(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error)
	updateFn  func(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error)
	deleteFn  func(ctx context.Context, id int64) error
}

func (m *mockSpecialProjectRepo) Create(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error) {
	if m.createFn != nil {
		return m.createFn(ctx, proj)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) GetByID(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) List(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, statusFilter, searchQuery, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockSpecialProjectRepo) Update(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, update)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

var _ repository.SpecialProjectRepository = (*mockSpecialProjectRepo)(nil)

func TestSpecialProjectService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("empty title returns error", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Create(ctx, &models.SpecialProject{Title: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("already exists maps to ErrAlreadyExists", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			createFn: func(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error) {
				return nil, models.ErrSpecialProjectAlreadyExists
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Create(ctx, &models.SpecialProject{Title: "Duplicate"})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectAlreadyExists)
	})

	t.Run("success creates and returns db model", func(t *testing.T) {
		img := "https://example.com/img.jpg"
		repo := &mockSpecialProjectRepo{
			createFn: func(ctx context.Context, proj *models.SpecialProject) (*models.SpecialProjectDB, error) {
				return &models.SpecialProjectDB{
					ID:    1,
					Title: proj.Title,
					Image: proj.Image,
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.Create(ctx, &models.SpecialProject{
			Title:  "Test",
			Image:  &img,
			Status: "active",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Test", got.Title)
	})
}

func TestSpecialProjectService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			getByIDFn: func(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
				return nil, models.ErrSpecialProjectNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.GetByID(ctx, 999)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectNotFound)
	})

	t.Run("success returns db model", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			getByIDFn: func(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
				return &models.SpecialProjectDB{
					ID:        id,
					Title:     "T",
					Status:    "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "T", got.Title)
		assert.Equal(t, "active", got.Status)
	})
}

func TestSpecialProjectService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id <= 0", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		title := "T"
		_, err := svc.Update(ctx, 0, &models.SpecialProjectUpdate{Title: &title})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("nil update", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Update(ctx, 1, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			updateFn: func(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
				return nil, models.ErrSpecialProjectNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		title := "T"
		_, err := svc.Update(ctx, 1, &models.SpecialProjectUpdate{Title: &title})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectNotFound)
	})

	t.Run("success returns updated db model", func(t *testing.T) {
		title := "Updated"
		desc := "new desc"
		status := "inactive"
		repo := &mockSpecialProjectRepo{
			updateFn: func(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
				return &models.SpecialProjectDB{
					ID:          id,
					Title:       *update.Title,
					Description: *update.Description,
					Status:      *update.Status,
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.Update(ctx, 1, &models.SpecialProjectUpdate{
			Title:       &title,
			Description: &desc,
			Status:      &status,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Updated", got.Title)
		assert.Equal(t, "inactive", got.Status)
	})
}

func TestSpecialProjectService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id <= 0", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		err := svc.Delete(ctx, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			deleteFn: func(ctx context.Context, id int64) error {
				return models.ErrSpecialProjectNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		err := svc.Delete(ctx, 999)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectNotFound)
	})

	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockSpecialProjectRepo{
			deleteFn: func(ctx context.Context, id int64) error {
				called = true
				assert.Equal(t, int64(1), id)
				return nil
			},
		}
		svc := NewSpecialProjectService(repo)
		err := svc.Delete(ctx, 1)
		require.NoError(t, err)
		assert.True(t, called)
	})
}

func TestSpecialProjectService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns list", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			listFn: func(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error) {
				return []*models.SpecialProjectDB{
					{ID: 1, Title: "A", Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: 2, Title: "B", Status: "inactive", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}, 10, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, total, err := svc.List(ctx, "active", "q", 10, 0)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, 10, total)
		assert.Equal(t, int64(1), got[0].ID)
		assert.Equal(t, "active", got[0].Status)
	})

	t.Run("repo error propagated", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockSpecialProjectRepo{
			listFn: func(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error) {
				return nil, 0, repoErr
			},
		}
		svc := NewSpecialProjectService(repo)
		_, _, err := svc.List(ctx, "", "", 0, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
	})
}
