//go:build integration

package postgres

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

func TestApplicationRepo_GetApplicationByID(t *testing.T) {
	appRepo := NewApplicationRepository(db)
	ctx := context.Background()

	created, err := appRepo.CreateApplication(ctx, &models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeBox,
		Source:       models.ApplicationSourceManual,
		CustomerName: "GetByID User",
		ContactInfo:  "getbyid@example.com",
	})
	require.NoError(t, err)

	t.Run("happy path — existing id", func(t *testing.T) {
		app, err := appRepo.GetApplicationByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, app.ID)
		assert.Equal(t, models.ApplicationTypeBox, app.Type)
		assert.Equal(t, "GetByID User", app.CustomerName)
		assert.Equal(t, "getbyid@example.com", app.ContactInfo)
	})

	t.Run("not found — unknown id", func(t *testing.T) {
		_, err := appRepo.GetApplicationByID(ctx, -1)
		assert.ErrorIs(t, err, models.ErrApplicationNotFound)
	})
}

func TestApplicationRepo_UpdateApplication(t *testing.T) {
	appRepo := NewApplicationRepository(db)
	ctx := context.Background()

	created, err := appRepo.CreateApplication(ctx, &models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeSpecialProject,
		Source:       models.ApplicationSourceManual,
		CustomerName: "Update User",
		ContactInfo:  "update@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, models.ApplicationStatusQueue, created.Status)

	t.Run("DoD: status changes and is visible in list", func(t *testing.T) {
		newStatus := models.ApplicationStatusInProgress
		updated, err := appRepo.UpdateApplication(ctx, created.ID, &models.ApplicationUpdateRequest{
			Status: &newStatus,
		})
		require.NoError(t, err)
		assert.Equal(t, models.ApplicationStatusInProgress, updated.Status)
		assert.Equal(t, created.ID, updated.ID)

		fetched, err := appRepo.GetApplicationByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, models.ApplicationStatusInProgress, fetched.Status)
	})

	t.Run("DoD: contact_info update", func(t *testing.T) {
		newContact := "updated@example.com"
		updated, err := appRepo.UpdateApplication(ctx, created.ID, &models.ApplicationUpdateRequest{
			ContactInfo: &newContact,
		})
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", updated.ContactInfo)
	})

	t.Run("DoD: box_id update visible in card", func(t *testing.T) {
		boxID := int64(1)
		updated, err := appRepo.UpdateApplication(ctx, created.ID, &models.ApplicationUpdateRequest{
			BoxID: &boxID,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), *updated.BoxID)

		fetched, err := appRepo.GetApplicationByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), *fetched.BoxID)
	})

	t.Run("not found — unknown id", func(t *testing.T) {
		newStatus := models.ApplicationStatusDone
		_, err := appRepo.UpdateApplication(ctx, -1, &models.ApplicationUpdateRequest{
			Status: &newStatus,
		})
		assert.ErrorIs(t, err, models.ErrApplicationNotFound)
	})

	t.Run("invalid input — no fields", func(t *testing.T) {
		_, err := appRepo.UpdateApplication(ctx, created.ID, &models.ApplicationUpdateRequest{})
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("invalid input — nil request", func(t *testing.T) {
		_, err := appRepo.UpdateApplication(ctx, created.ID, nil)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})
}

func TestApplicationRepo_DeleteApplication(t *testing.T) {
	appRepo := NewApplicationRepository(db)
	ctx := context.Background()

	created, err := appRepo.CreateApplication(ctx, &models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeBox,
		Source:       models.ApplicationSourceTelegramBot,
		CustomerName: "Delete User",
		ContactInfo:  "delete@example.com",
	})
	require.NoError(t, err)

	t.Run("happy path — existing application deleted", func(t *testing.T) {
		err := appRepo.DeleteApplication(ctx, created.ID)
		require.NoError(t, err)

		_, err = appRepo.GetApplicationByID(ctx, created.ID)
		assert.ErrorIs(t, err, models.ErrApplicationNotFound)
	})

	t.Run("not found — already deleted", func(t *testing.T) {
		err := appRepo.DeleteApplication(ctx, created.ID)
		assert.ErrorIs(t, err, models.ErrApplicationNotFound)
	})

	t.Run("not found — unknown id", func(t *testing.T) {
		err := appRepo.DeleteApplication(ctx, -1)
		assert.ErrorIs(t, err, models.ErrApplicationNotFound)
	})
}
