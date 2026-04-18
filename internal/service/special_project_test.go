package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository"
)

type mockSpecialProjectRepo struct {
	createFn  func(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error)
	getByIDFn func(ctx context.Context, id int64) (*models.SpecialProjectDB, error)
	listFn    func(ctx context.Context, statusFilter string, searchQuery string, limit, offset int) ([]*models.SpecialProjectDB, int, error)
	updateFn  func(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error)
	deleteFn  func(ctx context.Context, id int64) error
}

func (m *mockSpecialProjectRepo) Create(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
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

// Ensure mock implements the interface.
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
			createFn: func(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
				return nil, models.ErrSpecialProjectAlreadyExists
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Create(ctx, &models.SpecialProject{Title: "Duplicate"})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectAlreadyExists)
	})

	t.Run("success creates and returns domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			createFn: func(ctx context.Context, proj *models.SpecialProjectDB) (*models.SpecialProjectDB, error) {
				proj.ID = 1
				return proj, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		desc := "desc"
		got, err := svc.Create(ctx, &models.SpecialProject{
			Title:       "Test",
			Description: &desc,
			Image:       "img",
			Status:      "active",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Test", got.Title)
		assert.Equal(t, models.ServiceStatus("active"), got.Status)
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

	t.Run("success returns domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			getByIDFn: func(ctx context.Context, id int64) (*models.SpecialProjectDB, error) {
				return &models.SpecialProjectDB{ID: id, Title: "T", Status: "active"}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "T", got.Title)
		assert.Equal(t, models.ServiceStatus("active"), got.Status)
	})
}

func TestSpecialProjectService_UpdateSpecialProject(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id <= 0", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Update(ctx, 0, &models.SpecialProject{Title: "T"})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("nil project", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Update(ctx, 1, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("title too long", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		longTitle := strings.Repeat("x", 256)
		_, err := svc.Update(ctx, 1, &models.SpecialProject{Title: longTitle})
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
		_, err := svc.Update(ctx, 1, &models.SpecialProject{Title: "T"})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrSpecialProjectNotFound)
	})

	t.Run("success returns updated domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			updateFn: func(ctx context.Context, id int64, update *models.SpecialProjectUpdate) (*models.SpecialProjectDB, error) {
				return &models.SpecialProjectDB{
					ID: id, Title: update.Title, Description: update.Description,
					Image: update.Image, Status: update.Status,
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		desc := "new desc"
		got, err := svc.Update(ctx, 1, &models.SpecialProject{
			Title: "Updated", Description: &desc, Image: "img2", Status: "inactive",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Updated", got.Title)
		assert.Equal(t, models.ServiceStatus("inactive"), got.Status)
	})
}

func TestSpecialProjectService_DeleteSpecialProject(t *testing.T) {
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
					{ID: 1, Title: "A", Status: "active"},
					{ID: 2, Title: "B", Status: "inactive"},
				}, 10, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, _, err := svc.List(ctx, "active", "q", 10, 0)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, int64(1), got[0].ID)
		assert.Equal(t, "A", got[0].Title)
		assert.Equal(t, models.ServiceStatus("active"), got[0].Status)
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
