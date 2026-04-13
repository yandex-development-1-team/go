package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-development-1-team/go/internal/models"
)

func TestResourcePageRepo_GetAll(t *testing.T) {
	ctx := context.Background()

	pages, err := resourcePageRepo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, pages, 4)
}

func TestResourcePageRepo_GetBySlug(t *testing.T) {
	ctx := context.Background()

	t.Run("existing slug", func(t *testing.T) {
		page, err := resourcePageRepo.GetBySlug(ctx, "faq")
		require.NoError(t, err)
		assert.Equal(t, "faq", page.Slug)
		assert.NotEmpty(t, page.Title)
		assert.NotNil(t, page.Links)
	})
}

func TestResourcePageRepo_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("updates content", func(t *testing.T) {
		page := models.ResourcePage{
			Title:   "Информация об организации",
			Content: "Новое описание",
			Links:   []models.ResourcePageLink{},
		}

		updated, err := resourcePageRepo.Update(ctx, "org-info", page)
		require.NoError(t, err)
		assert.Equal(t, "Новое описание", updated.Content)
		assert.Empty(t, updated.Links)
	})

	t.Run("generates uuid for new link", func(t *testing.T) {
		page := models.ResourcePage{
			Title:   "Полезные ссылки",
			Content: "",
			Links: []models.ResourcePageLink{
				{Title: "Сайт", URL: "https://example.com"},
			},
		}

		updated, err := resourcePageRepo.Update(ctx, "useful-links", page)
		require.NoError(t, err)
		assert.Len(t, updated.Links, 1)
		assert.NotEmpty(t, updated.Links[0].ID)
		assert.Equal(t, "Сайт", updated.Links[0].Title)
	})

	t.Run("keeps existing uuid", func(t *testing.T) {
		existingID := "550e8400-e29b-41d4-a716-446655440000"
		page := models.ResourcePage{
			Title:   "Полезные ссылки",
			Content: "",
			Links: []models.ResourcePageLink{
				{ID: existingID, Title: "Сайт", URL: "https://example.com"},
			},
		}

		updated, err := resourcePageRepo.Update(ctx, "useful-links", page)
		require.NoError(t, err)
		assert.Equal(t, existingID, updated.Links[0].ID)
	})
}

func TestResourcePageRepo_Clear(t *testing.T) {
	ctx := context.Background()

	t.Run("clears content and links", func(t *testing.T) {
		_, err := resourcePageRepo.Update(ctx, "faq", models.ResourcePage{
			Title:   "FAQ",
			Content: "Какой-то контент",
			Links:   []models.ResourcePageLink{{Title: "Ссылка", URL: "https://example.com"}},
		})
		require.NoError(t, err)

		cleared, err := resourcePageRepo.Clear(ctx, "faq")
		require.NoError(t, err)
		assert.Equal(t, "faq", cleared.Slug)
		assert.Empty(t, cleared.Content)
		assert.Empty(t, cleared.Links)
	})
}

func TestResourcePageRepo_DeleteLink(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes link by id", func(t *testing.T) {
		page := models.ResourcePage{
			Title:   "Полезные ссылки",
			Content: "",
			Links: []models.ResourcePageLink{
				{Title: "Первая", URL: "https://first.com"},
				{Title: "Вторая", URL: "https://second.com"},
			},
		}
		updated, err := resourcePageRepo.Update(ctx, "useful-links", page)
		require.NoError(t, err)
		require.Len(t, updated.Links, 2)

		idToDelete := updated.Links[0].ID

		result, err := resourcePageRepo.DeleteLink(ctx, "useful-links", idToDelete)
		require.NoError(t, err)
		assert.Len(t, result.Links, 1)
		assert.Equal(t, updated.Links[1].ID, result.Links[0].ID)
	})

	t.Run("returns empty links when last link deleted", func(t *testing.T) {
		page := models.ResourcePage{
			Title:   "Полезные ссылки",
			Content: "",
			Links: []models.ResourcePageLink{
				{Title: "Единственная", URL: "https://example.com"},
			},
		}
		updated, err := resourcePageRepo.Update(ctx, "useful-links", page)
		require.NoError(t, err)
		require.Len(t, updated.Links, 1)

		result, err := resourcePageRepo.DeleteLink(ctx, "useful-links", updated.Links[0].ID)
		require.NoError(t, err)
		assert.Empty(t, result.Links)
	})
}
