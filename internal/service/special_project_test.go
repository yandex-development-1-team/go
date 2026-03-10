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
	"github.com/yandex-development-1-team/go/internal/specialproject"
)

type mockSpecialProjectRepo struct {
	createFn               func(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error)
	getByIDFn              func(ctx context.Context, id int64) (*specialproject.DB, error)
	listFn                 func(ctx context.Context, statusFilter *bool, searchQuery string) ([]*specialproject.DB, error)
	updateSpecialProjectFn func(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error)
	deleteSpecialProjectFn func(ctx context.Context, id int64) error
}

func (m *mockSpecialProjectRepo) Create(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error) {
	if m.createFn != nil {
		return m.createFn(ctx, proj)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) GetByID(ctx context.Context, id int64) (*specialproject.DB, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) List(ctx context.Context, statusFilter *bool, searchQuery string) ([]*specialproject.DB, error) {
	if m.listFn != nil {
		return m.listFn(ctx, statusFilter, searchQuery)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) UpdateSpecialProject(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error) {
	if m.updateSpecialProjectFn != nil {
		return m.updateSpecialProjectFn(ctx, id, update)
	}
	return nil, nil
}

func (m *mockSpecialProjectRepo) DeleteSpecialProject(ctx context.Context, id int64) error {
	if m.deleteSpecialProjectFn != nil {
		return m.deleteSpecialProjectFn(ctx, id)
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
		_, err := svc.Create(ctx, &specialproject.Project{Title: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("already exists maps to ErrAlreadyExists", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			createFn: func(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error) {
				return nil, specialproject.ErrAlreadyExists
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.Create(ctx, &specialproject.Project{Title: "Duplicate"})
		require.Error(t, err)
		assert.ErrorIs(t, err, specialproject.ErrAlreadyExists)
	})

	t.Run("success creates and returns domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			createFn: func(ctx context.Context, proj *specialproject.DB) (*specialproject.DB, error) {
				proj.ID = 1
				return proj, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		desc := "desc"
		got, err := svc.Create(ctx, &specialproject.Project{
			Title:         "Test",
			Description:   &desc,
			Image:         "img",
			IsActiveInBot: true,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Test", got.Title)
		assert.True(t, got.IsActiveInBot)
	})
}

func TestSpecialProjectService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			getByIDFn: func(ctx context.Context, id int64) (*specialproject.DB, error) {
				return nil, specialproject.ErrNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.GetByID(ctx, 999)
		require.Error(t, err)
		assert.ErrorIs(t, err, specialproject.ErrNotFound)
	})

	t.Run("success returns domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			getByIDFn: func(ctx context.Context, id int64) (*specialproject.DB, error) {
				return &specialproject.DB{ID: id, Title: "T", IsActiveInBot: true}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.GetByID(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "T", got.Title)
		assert.True(t, got.IsActiveInBot)
	})
}

func TestSpecialProjectService_UpdateSpecialProject(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id <= 0", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.UpdateSpecialProject(ctx, 0, &specialproject.Project{Title: "T"})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("nil project", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		_, err := svc.UpdateSpecialProject(ctx, 1, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("title too long", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		longTitle := strings.Repeat("x", 256)
		_, err := svc.UpdateSpecialProject(ctx, 1, &specialproject.Project{Title: longTitle})
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			updateSpecialProjectFn: func(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error) {
				return nil, specialproject.ErrNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.UpdateSpecialProject(ctx, 1, &specialproject.Project{Title: "T"})
		require.Error(t, err)
		assert.ErrorIs(t, err, specialproject.ErrNotFound)
	})

	t.Run("success returns updated domain", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			updateSpecialProjectFn: func(ctx context.Context, id int64, update *specialproject.Update) (*specialproject.DB, error) {
				return &specialproject.DB{
					ID: id, Title: update.Title, Description: update.Description,
					Image: update.Image, IsActiveInBot: update.IsActiveInBot,
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		desc := "new desc"
		got, err := svc.UpdateSpecialProject(ctx, 1, &specialproject.Project{
			Title: "Updated", Description: &desc, Image: "img2", IsActiveInBot: false,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), got.ID)
		assert.Equal(t, "Updated", got.Title)
		assert.False(t, got.IsActiveInBot)
	})
}

func TestSpecialProjectService_DeleteSpecialProject(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid id <= 0", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{}
		svc := NewSpecialProjectService(repo)
		err := svc.DeleteSpecialProject(ctx, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("not found maps to ErrNotFound", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			deleteSpecialProjectFn: func(ctx context.Context, id int64) error {
				return specialproject.ErrNotFound
			},
		}
		svc := NewSpecialProjectService(repo)
		err := svc.DeleteSpecialProject(ctx, 999)
		require.Error(t, err)
		assert.ErrorIs(t, err, specialproject.ErrNotFound)
	})

	t.Run("success", func(t *testing.T) {
		called := false
		repo := &mockSpecialProjectRepo{
			deleteSpecialProjectFn: func(ctx context.Context, id int64) error {
				called = true
				assert.Equal(t, int64(1), id)
				return nil
			},
		}
		svc := NewSpecialProjectService(repo)
		err := svc.DeleteSpecialProject(ctx, 1)
		require.NoError(t, err)
		assert.True(t, called)
	})
}

func TestSpecialProjectService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("success returns list", func(t *testing.T) {
		repo := &mockSpecialProjectRepo{
			listFn: func(ctx context.Context, statusFilter *bool, searchQuery string) ([]*specialproject.DB, error) {
				return []*specialproject.DB{
					{ID: 1, Title: "A", IsActiveInBot: true},
					{ID: 2, Title: "B", IsActiveInBot: false},
				}, nil
			},
		}
		svc := NewSpecialProjectService(repo)
		got, err := svc.List(ctx, "active", "q")
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, int64(1), got[0].ID)
		assert.Equal(t, "A", got[0].Title)
		assert.True(t, got[0].IsActiveInBot)
	})

	t.Run("repo error propagated", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockSpecialProjectRepo{
			listFn: func(ctx context.Context, statusFilter *bool, searchQuery string) ([]*specialproject.DB, error) {
				return nil, repoErr
			},
		}
		svc := NewSpecialProjectService(repo)
		_, err := svc.List(ctx, "", "")
		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
	})
}
