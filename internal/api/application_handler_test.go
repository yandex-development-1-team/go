package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
)

// mockApplicationRepo implements repository.ApplicationRepository for unit tests.
type mockApplicationRepo struct {
	createFn      func(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error)
	listFn        func(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error)
	getByIDFn     func(ctx context.Context, id int64) (*models.Application, error)
	updateFn      func(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error)
	deleteFn      func(ctx context.Context, id int64) error
}

func (m *mockApplicationRepo) CreateApplication(ctx context.Context, req *models.ApplicationCreateRequest) (*models.Application, error) {
	return m.createFn(ctx, req)
}
func (m *mockApplicationRepo) GetApplications(ctx context.Context, filter models.ApplicationFilter) ([]models.Application, int, error) {
	return m.listFn(ctx, filter)
}
func (m *mockApplicationRepo) GetApplicationByID(ctx context.Context, id int64) (*models.Application, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockApplicationRepo) UpdateApplication(ctx context.Context, id int64, req *models.ApplicationUpdateRequest) (*models.Application, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockApplicationRepo) DeleteApplication(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func newHandlerMux(repo *mockApplicationRepo) *http.ServeMux {
	mux := http.NewServeMux()
	NewApplicationHandler(repo).RegisterRoutes(mux)
	return mux
}

func sampleApp() *models.Application {
	return &models.Application{
		ID:          42,
		Type:        models.ApplicationTypeBox,
		Source:      models.ApplicationSourceManual,
		Status:      models.ApplicationStatusQueue,
		CustomerName: "Иван",
		ContactInfo: "ivan@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── create ────────────────────────────────────────────────────────────────────

func TestCreate_HappyPath(t *testing.T) {
	repo := &mockApplicationRepo{
		createFn: func(_ context.Context, _ *models.ApplicationCreateRequest) (*models.Application, error) {
			return sampleApp(), nil
		},
	}
	body, _ := json.Marshal(models.ApplicationCreateRequest{
		Type:         models.ApplicationTypeBox,
		Source:       models.ApplicationSourceManual,
		CustomerName: "Иван",
		ContactInfo:  "ivan@example.com",
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreate_ValidationError(t *testing.T) {
	repo := &mockApplicationRepo{}
	body, _ := json.Marshal(models.ApplicationCreateRequest{
		Type:   models.ApplicationTypeBox,
		Source: models.ApplicationSourceManual,
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/applications", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── list ─────────────────────────────────────────────────────────────────────

func TestList_HappyPath(t *testing.T) {
	repo := &mockApplicationRepo{
		listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
			return []models.Application{*sampleApp()}, 1, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.ApplicationListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Pagination.Total)
	assert.Len(t, resp.Items, 1)
}

func TestList_EmptyReturnsEmptySlice(t *testing.T) {
	repo := &mockApplicationRepo{
		listFn: func(_ context.Context, _ models.ApplicationFilter) ([]models.Application, int, error) {
			return nil, 0, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.ApplicationListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Items)
	assert.Len(t, resp.Items, 0)
}

// ── getByID ───────────────────────────────────────────────────────────────────

func TestGetByID_HappyPath(t *testing.T) {
	repo := &mockApplicationRepo{
		getByIDFn: func(_ context.Context, id int64) (*models.Application, error) {
			app := sampleApp()
			app.ID = id
			return app, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications/42", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var app models.Application
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &app))
	assert.Equal(t, int64(42), app.ID)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockApplicationRepo{
		getByIDFn: func(_ context.Context, _ int64) (*models.Application, error) {
			return nil, models.ErrApplicationNotFound
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications/999", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetByID_InvalidID(t *testing.T) {
	repo := &mockApplicationRepo{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications/abc", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetByID_ZeroID(t *testing.T) {
	repo := &mockApplicationRepo{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/applications/0", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── update ────────────────────────────────────────────────────────────────────

func TestUpdate_HappyPath(t *testing.T) {
	status := models.ApplicationStatusInProgress
	repo := &mockApplicationRepo{
		updateFn: func(_ context.Context, _ int64, _ *models.ApplicationUpdateRequest) (*models.Application, error) {
			app := sampleApp()
			app.Status = status
			return app, nil
		},
	}
	body, _ := json.Marshal(models.ApplicationUpdateRequest{Status: &status})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/42", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var app models.Application
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &app))
	assert.Equal(t, models.ApplicationStatusInProgress, app.Status)
}

func TestUpdate_NotFound(t *testing.T) {
	status := models.ApplicationStatusDone
	repo := &mockApplicationRepo{
		updateFn: func(_ context.Context, _ int64, _ *models.ApplicationUpdateRequest) (*models.Application, error) {
			return nil, models.ErrApplicationNotFound
		},
	}
	body, _ := json.Marshal(models.ApplicationUpdateRequest{Status: &status})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/999", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdate_NoFields(t *testing.T) {
	repo := &mockApplicationRepo{}
	body, _ := json.Marshal(models.ApplicationUpdateRequest{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/42", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdate_InvalidStatus(t *testing.T) {
	repo := &mockApplicationRepo{}
	invalidStatus := models.ApplicationStatus("unknown")
	body, _ := json.Marshal(models.ApplicationUpdateRequest{Status: &invalidStatus})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/42", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdate_InvalidID(t *testing.T) {
	repo := &mockApplicationRepo{}
	body, _ := json.Marshal(models.ApplicationUpdateRequest{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/abc", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdate_AllowedFields(t *testing.T) {
	contact := "new@example.com"
	boxID := int64(7)
	repo := &mockApplicationRepo{
		updateFn: func(_ context.Context, _ int64, req *models.ApplicationUpdateRequest) (*models.Application, error) {
			app := sampleApp()
			app.ContactInfo = *req.ContactInfo
			app.BoxID = req.BoxID
			return app, nil
		},
	}
	body, _ := json.Marshal(models.ApplicationUpdateRequest{ContactInfo: &contact, BoxID: &boxID})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/applications/42", bytes.NewReader(body))
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	var app models.Application
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &app))
	assert.Equal(t, "new@example.com", app.ContactInfo)
	assert.Equal(t, int64(7), *app.BoxID)
}

// ── delete ────────────────────────────────────────────────────────────────────

func TestDelete_HappyPath(t *testing.T) {
	repo := &mockApplicationRepo{
		deleteFn: func(_ context.Context, _ int64) error {
			return nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/42", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "message")
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockApplicationRepo{
		deleteFn: func(_ context.Context, _ int64) error {
			return models.ErrApplicationNotFound
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/999", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDelete_InvalidID(t *testing.T) {
	repo := &mockApplicationRepo{}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/applications/abc", nil)
	newHandlerMux(repo).ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
