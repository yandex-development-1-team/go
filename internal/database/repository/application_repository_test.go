//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
)

func TestApplicationRepo_CreateApplication(t *testing.T) {
	appRepo := NewApplicationRepository(db)
	ctx := context.Background()

	t.Run("happy path — bot application", func(t *testing.T) {
		req := &models.ApplicationCreateRequest{
			Type:         models.ApplicationTypeBox,
			Source:       models.ApplicationSourceTelegramBot,
			CustomerName: "Иван Иванов",
			ContactInfo:  "ivan@example.com",
		}

		app, err := appRepo.CreateApplication(ctx, req)

		require.NoError(t, err)
		assert.Greater(t, app.ID, int64(0))
		assert.Equal(t, models.ApplicationTypeBox, app.Type)
		assert.Equal(t, models.ApplicationSourceTelegramBot, app.Source)
		assert.Equal(t, models.ApplicationStatusQueue, app.Status)
		assert.Equal(t, "Иван Иванов", app.CustomerName)
		assert.Equal(t, "ivan@example.com", app.ContactInfo)
		assert.False(t, app.CreatedAt.IsZero())
	})

	t.Run("invalid input — empty customer_name", func(t *testing.T) {
		req := &models.ApplicationCreateRequest{
			Type:         models.ApplicationTypeBox,
			Source:       models.ApplicationSourceManual,
			CustomerName: "",
			ContactInfo:  "test@example.com",
		}
		_, err := appRepo.CreateApplication(ctx, req)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("invalid input — bad type enum", func(t *testing.T) {
		req := &models.ApplicationCreateRequest{
			Type:         "unknown_type",
			Source:       models.ApplicationSourceManual,
			CustomerName: "Test",
			ContactInfo:  "test@example.com",
		}
		_, err := appRepo.CreateApplication(ctx, req)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})
}

func TestApplicationRepo_GetApplications(t *testing.T) {
	appRepo := NewApplicationRepository(db)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `DELETE FROM applications`)
	require.NoError(t, err)

	for i := 0; i < 2; i++ {
		_, err := appRepo.CreateApplication(ctx, &models.ApplicationCreateRequest{
			Type:         models.ApplicationTypeBox,
			Source:       models.ApplicationSourceTelegramBot,
			CustomerName: "Клиент box",
			ContactInfo:  "box@example.com",
		})
		require.NoError(t, err)
	}
	_, err = appRepo.CreateApplication(ctx, &models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeSpecialProject,
		Source:       models.ApplicationSourceManual,
		CustomerName: "Клиент special",
		ContactInfo:  "sp@example.com",
	})
	require.NoError(t, err)

	t.Run("list all — total=3", func(t *testing.T) {
		apps, total, err := appRepo.GetApplications(ctx, models.ApplicationFilter{Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, apps, 3)
	})

	t.Run("filter by type=box — total=2", func(t *testing.T) {
		appType := models.ApplicationTypeBox
		apps, total, err := appRepo.GetApplications(ctx, models.ApplicationFilter{Type: &appType, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, apps, 2)
		for _, a := range apps {
			assert.Equal(t, models.ApplicationTypeBox, a.Type)
		}
	})

	t.Run("pagination limit=1 offset=0", func(t *testing.T) {
		apps, total, err := appRepo.GetApplications(ctx, models.ApplicationFilter{Limit: 1, Offset: 0})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, apps, 1)
	})

	t.Run("pagination limit=1 offset=2", func(t *testing.T) {
		apps, total, err := appRepo.GetApplications(ctx, models.ApplicationFilter{Limit: 1, Offset: 2})
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, apps, 1)
	})

	t.Run("DoD: bot application visible in filtered list", func(t *testing.T) {
		appType := models.ApplicationTypeBox
		appStatus := models.ApplicationStatusQueue
		apps, total, err := appRepo.GetApplications(ctx, models.ApplicationFilter{
			Type: &appType, Status: &appStatus, Limit: 20,
		})
		require.NoError(t, err)
		assert.Greater(t, total, 0)
		for _, a := range apps {
			assert.Equal(t, models.ApplicationTypeBox, a.Type)
			assert.Equal(t, models.ApplicationStatusQueue, a.Status)
		}
	})
}
