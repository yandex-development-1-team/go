package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/service/api/mocks"

	"go.uber.org/mock/gomock"
)

var fakeServices = []models.Service{
	{
		ID:          1,
		Name:        "Test Box",
		Status:      "active",
		Price:       1000,
		Location:    "Москва",
		Organizer:   "Иван",
		Description: "Длинное описание которое может не влезть в одну строку и должно переноситься",
		Rules:       "Правила участия",
		BoxAvailableSlots: []models.BoxAvailableSlot{
			{Date: "2024-03-25", StartTime: "09:00", EndTime: "18:00"},
			{Date: "2024-03-26", StartTime: "10:00", EndTime: "15:00"},
		},
	},
	{
		ID:        2,
		Name:      "Test Box 2",
		Status:    "active",
		Price:     2000,
		Location:  "СПб",
		Organizer: "Петр",
	},
}

func TestUpdate(t *testing.T) {
	serviceID := int64(1)
	newName := "Updated Box"

	expectedService := &models.Service{
		ID:   serviceID,
		Name: newName,
	}

	t.Run("success - update fields only, no slots", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Name: &newName,
		}

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			GetServiceByID(gomock.Any(), serviceID).
			Return(expectedService, nil)

		result, err := svc.Update(context.Background(), serviceID, req)
		require.NoError(t, err)
		assert.Equal(t, expectedService, result)
	})

	t.Run("success - update with slots", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Name: &newName,
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "09:00", EndTime: "18:00"},
				{Date: "2024-03-26", StartTime: "10:00", EndTime: "15:00"},
			},
		}

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			DeleteServiceSlots(gomock.Any(), serviceID).
			Return(nil)

		mockLister.EXPECT().
			UpdateServiceSlots(gomock.Any(), serviceID, gomock.Any()).
			Return(nil)

		mockLister.EXPECT().
			GetServiceByID(gomock.Any(), serviceID).
			Return(expectedService, nil)

		result, err := svc.Update(context.Background(), serviceID, req)
		require.NoError(t, err)
		assert.Equal(t, expectedService, result)
	})

	t.Run("success - empty slots, delete all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Name:  &newName,
			Slots: []models.BoxAvailableSlot{}, // пустой — удаляем все слоты
		}

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			DeleteServiceSlots(gomock.Any(), serviceID).
			Return(nil)

		// len(Slots) == 0 — UpdateServiceSlots не вызывается

		mockLister.EXPECT().
			GetServiceByID(gomock.Any(), serviceID).
			Return(expectedService, nil)

		result, err := svc.Update(context.Background(), serviceID, req)
		require.NoError(t, err)
		assert.Equal(t, expectedService, result)
	})

	t.Run("invalid slot date", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "invalid-date", StartTime: "09:00", EndTime: "18:00"},
			},
		}

		// Ни один метод не должен вызываться
		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("invalid slot start time", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "invalid", EndTime: "18:00"},
			},
		}

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("invalid slot end time", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "09:00", EndTime: "invalid"},
			},
		}

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, models.ErrInvalidInput)
	})

	t.Run("UpdateService error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{Name: &newName}
		dbErr := errors.New("db error")

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(dbErr)

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("DeleteServiceSlots error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "09:00", EndTime: "18:00"},
			},
		}
		dbErr := errors.New("delete error")

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			DeleteServiceSlots(gomock.Any(), serviceID).
			Return(dbErr)

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("UpdateServiceSlots error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "09:00", EndTime: "18:00"},
			},
		}
		dbErr := errors.New("insert error")

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			DeleteServiceSlots(gomock.Any(), serviceID).
			Return(nil)

		mockLister.EXPECT().
			UpdateServiceSlots(gomock.Any(), serviceID, gomock.Any()).
			Return(dbErr)

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("GetServiceByID error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{Name: &newName}

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			GetServiceByID(gomock.Any(), serviceID).
			Return(nil, models.ErrBoxSolutionNotFound)

		result, err := svc.Update(context.Background(), serviceID, req)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, models.ErrBoxSolutionNotFound)
	})

	t.Run("slots parsed correctly", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
		mockTxRepo := mocks.NewMockTxRepository(ctrl)
		svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

		req := &models.BoxUpdate{
			Slots: []models.BoxAvailableSlot{
				{Date: "2024-03-25", StartTime: "09:00", EndTime: "18:00"},
			},
		}

		expectedSlots := &models.BoxNewSlots{
			Date:      []time.Time{time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC)},
			StartTime: []time.Time{time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)},
			EndTime:   []time.Time{time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC)},
		}

		mockTxRepo.EXPECT().
			RunToTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		mockLister.EXPECT().
			UpdateService(gomock.Any(), serviceID, req).
			Return(nil)

		mockLister.EXPECT().
			DeleteServiceSlots(gomock.Any(), serviceID).
			Return(nil)

		mockLister.EXPECT().
			UpdateServiceSlots(gomock.Any(), serviceID, expectedSlots).
			Return(nil)

		mockLister.EXPECT().
			GetServiceByID(gomock.Any(), serviceID).
			Return(expectedService, nil)

		result, err := svc.Update(context.Background(), serviceID, req)
		require.NoError(t, err)
		assert.Equal(t, expectedService, result)
	})
}

func TestExport_PDF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMockBoxSolutionRepository(ctrl)
	activeStatus := models.ServiceStatus("active")

	mockLister.EXPECT().
		GetServicesByStatus(gomock.Any(), &activeStatus).
		Return(fakeServices, nil)

	mockTxRepo := mocks.NewMockTxRepository(ctrl)
	svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

	data, contentType, err := svc.Export(context.Background(), "active", "pdf")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/pdf" {
		t.Errorf("expected application/pdf, got %s", contentType)
	}
	if len(data) < 100 {
		t.Error("pdf too small, probably empty")
	}
	// Проверяем сигнатуру PDF
	if string(data[:4]) != "%PDF" {
		t.Error("data is not a valid PDF")
	}
}

func TestExport_CSV(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMockBoxSolutionRepository(ctrl)

	mockLister.EXPECT().
		GetServicesByStatus(gomock.Any(), nil).
		Return(fakeServices, nil)

	mockTxRepo := mocks.NewMockTxRepository(ctrl)
	svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

	data, contentType, err := svc.Export(context.Background(), "", "csv")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "text/csv" {
		t.Errorf("expected text/csv, got %s", contentType)
	}

	body := string(data)

	// Проверяем заголовки
	if !strings.Contains(body, "Название") {
		t.Error("csv missing header Название")
	}
	if !strings.Contains(body, "Организатор") {
		t.Error("csv missing header Организатор")
	}

	// Проверяем данные
	if !strings.Contains(body, "Test Box") {
		t.Error("csv missing service name")
	}
	if !strings.Contains(body, "2024-03-25") {
		t.Error("csv missing slot date")
	}

	// Проверяем что строк правильное количество
	// 1 заголовок + 2 слота box1 + 1 строка box2 без слотов = 4
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestExport_DefaultFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMockBoxSolutionRepository(ctrl)

	mockLister.EXPECT().
		GetServicesByStatus(gomock.Any(), nil).
		Return(fakeServices, nil)

	mockTxRepo := mocks.NewMockTxRepository(ctrl)
	svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

	// Передаём невалидный формат — должен вернуть pdf
	_, contentType, err := svc.Export(context.Background(), "", "xml")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/pdf" {
		t.Errorf("expected application/pdf for unknown format, got %s", contentType)
	}
}

func TestExport_ListerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMockBoxSolutionRepository(ctrl)

	mockLister.EXPECT().
		GetServicesByStatus(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db error"))

	mockTxRepo := mocks.NewMockTxRepository(ctrl)
	svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

	_, _, err := svc.Export(context.Background(), "active", "pdf")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExport_EmptyServices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMockBoxSolutionRepository(ctrl)

	mockLister.EXPECT().
		GetServicesByStatus(gomock.Any(), gomock.Any()).
		Return([]models.Service{}, nil)

	mockTxRepo := mocks.NewMockTxRepository(ctrl)
	svc := NewAPIBoxService(mockLister, nil, mockTxRepo)

	data, contentType, err := svc.Export(context.Background(), "", "pdf")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "application/pdf" {
		t.Errorf("expected application/pdf, got %s", contentType)
	}
	// Пустой PDF всё равно валидный
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Error("data is not a valid PDF")
	}
}
