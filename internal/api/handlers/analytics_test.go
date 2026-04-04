package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/dto"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

type mockExporter struct {
	result   apiService.ExportResult
	overview dto.AnalyticsOverview
	err      error
}

func (m *mockExporter) Export(_ context.Context, _ dto.AnalyticsExportRequest) (apiService.ExportResult, error) {
	return m.result, m.err
}

func (m *mockExporter) GetOverviewAnalytics(ctx context.Context, dateFrom *time.Time, dateTo *time.Time) (dto.AnalyticsOverview, error)

func (m *mockExporter) GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error)

func (m *mockExporter) GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error)

func (m *mockExporter) GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error)

type capturingExporter struct {
	capture *dto.AnalyticsExportRequest
}

func (c *capturingExporter) GetOverviewAnalytics(ctx context.Context, dateFrom *time.Time, dateTo *time.Time) (dto.AnalyticsOverview, error)

func (c *capturingExporter) GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error)

func (c *capturingExporter) GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error)

func (c *capturingExporter) GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error)

func (c *capturingExporter) Export(_ context.Context, req dto.AnalyticsExportRequest) (apiService.ExportResult, error) {
	*c.capture = req
	return apiService.ExportResult{Data: []byte("ok"), ContentType: "text/plain", Filename: "test.txt"}, nil
}

func newTestRouter(h *AnalyticsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/analytics/export", h.Export)
	return r
}

func exportRequest(url string) (*httptest.ResponseRecorder, *http.Request) {
	return httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, url, nil)
}

func TestAnalyticsHandler_Export_MissingType(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "errors")
}

func TestAnalyticsHandler_Export_InvalidType(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export?type=employees")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_Export_InvalidFormat(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export?type=boxes&format=pdf")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_Export_InvalidDateFrom(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export?type=boxes&date_from=01-01-2026")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "date_from")
}

func TestAnalyticsHandler_Export_InvalidDateTo(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export?type=boxes&date_to=not-a-date")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "date_to")
}

func TestAnalyticsHandler_Export_DateFromAfterDateTo(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w, req := exportRequest("/analytics/export?type=boxes&date_from=2026-03-01&date_to=2026-01-01")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "date_to")
}

func TestAnalyticsHandler_Export_BoxesXLSX_OK(t *testing.T) {
	xlsxData := []byte{0x50, 0x4B, 0x03, 0x04}
	h := &AnalyticsHandler{svc: &mockExporter{
		result: apiService.ExportResult{
			Data:        xlsxData,
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Filename:    "analytics_boxes.xlsx",
		},
	}}
	w, req := exportRequest("/analytics/export?type=boxes&format=xlsx")

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "analytics_boxes.xlsx")
	assert.Equal(t, xlsxData, w.Body.Bytes())
}

func TestAnalyticsHandler_Export_UsersCSV_OK(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{
		result: apiService.ExportResult{
			Data:        []byte("\xEF\xBB\xBFID,Email\n1,user@test.com\n"),
			ContentType: "text/csv; charset=utf-8",
			Filename:    "analytics_users.csv",
		},
	}}
	w, req := exportRequest("/analytics/export?type=users&format=csv")

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "analytics_users.csv")
}

func TestAnalyticsHandler_Export_DefaultFormatIsXLSX(t *testing.T) {
	var capturedReq dto.AnalyticsExportRequest
	h := &AnalyticsHandler{svc: &capturingExporter{capture: &capturedReq}}
	w, req := exportRequest("/analytics/export?type=boxes")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, dto.ExportFormatXLSX, capturedReq.Format)
}

func TestAnalyticsHandler_Export_ServiceError_Returns500(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{err: errors.New("unexpected db failure")}}
	w, req := exportRequest("/analytics/export?type=boxes")

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_Export_DateRange_Parsed(t *testing.T) {
	var capturedReq dto.AnalyticsExportRequest
	h := &AnalyticsHandler{svc: &capturingExporter{capture: &capturedReq}}
	w, req := exportRequest("/analytics/export?type=users&date_from=2026-01-01&date_to=2026-03-01")

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedReq.DateFrom)
	require.NotNil(t, capturedReq.DateTo)
	assert.Equal(t, "2026-01-01", capturedReq.DateFrom.Format("2006-01-02"))
	assert.Equal(t, "2026-03-01", capturedReq.DateTo.Format("2006-01-02"))
}
