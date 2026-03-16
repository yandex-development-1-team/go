package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/dto"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

// mockExporter is a test double for the analyticsExporter interface.
type mockExporter struct {
	result apiService.ExportResult
	err    error
}

func (m *mockExporter) Export(_ context.Context, _ dto.AnalyticsExportRequest) (apiService.ExportResult, error) {
	return m.result, m.err
}

func newTestRouter(h *AnalyticsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/analytics/export", h.Export)
	return r
}

func TestAnalyticsHandler_Export_MissingType(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "errors")
}

func TestAnalyticsHandler_Export_InvalidType(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=employees", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_Export_InvalidFormat(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes&format=pdf", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalyticsHandler_Export_InvalidDateFrom(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes&date_from=01-01-2026", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "date_from")
}

func TestAnalyticsHandler_Export_InvalidDateTo(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes&date_to=not-a-date", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "date_to")
}

func TestAnalyticsHandler_Export_BoxesXLSX_OK(t *testing.T) {
	xlsxData := []byte{0x50, 0x4B, 0x03, 0x04} // ZIP/XLSX magic bytes
	h := &AnalyticsHandler{svc: &mockExporter{
		result: apiService.ExportResult{
			Data:        xlsxData,
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Filename:    "analytics_boxes.xlsx",
		},
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes&format=xlsx", nil)

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "analytics_boxes.xlsx")
	assert.Equal(t, xlsxData, w.Body.Bytes())
}

func TestAnalyticsHandler_Export_UsersCSV_OK(t *testing.T) {
	csvData := []byte("\xEF\xBB\xBFID,Email\n1,user@test.com\n")
	h := &AnalyticsHandler{svc: &mockExporter{
		result: apiService.ExportResult{
			Data:        csvData,
			ContentType: "text/csv; charset=utf-8",
			Filename:    "analytics_users.csv",
		},
	}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=users&format=csv", nil)

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "analytics_users.csv")
}

func TestAnalyticsHandler_Export_DefaultFormatIsXLSX(t *testing.T) {
	var capturedReq dto.AnalyticsExportRequest
	h := &AnalyticsHandler{svc: &capturingExporter{capture: &capturedReq}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, dto.ExportFormatXLSX, capturedReq.Format)
}

func TestAnalyticsHandler_Export_ServiceError_Returns500(t *testing.T) {
	h := &AnalyticsHandler{svc: &mockExporter{err: errors.New("unexpected db failure")}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=boxes", nil)

	newTestRouter(h).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_Export_DateRange_Parsed(t *testing.T) {
	var capturedReq dto.AnalyticsExportRequest
	h := &AnalyticsHandler{svc: &capturingExporter{capture: &capturedReq}}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/analytics/export?type=users&date_from=2026-01-01&date_to=2026-03-01", nil)

	newTestRouter(h).ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedReq.DateFrom)
	require.NotNil(t, capturedReq.DateTo)
	assert.Equal(t, "2026-01-01", capturedReq.DateFrom.Format("2006-01-02"))
	assert.Equal(t, "2026-03-01", capturedReq.DateTo.Format("2006-01-02"))
}

// capturingExporter records the request it receives to allow assertion on parsed params.
type capturingExporter struct {
	capture *dto.AnalyticsExportRequest
}

func (c *capturingExporter) Export(_ context.Context, req dto.AnalyticsExportRequest) (apiService.ExportResult, error) {
	*c.capture = req
	return apiService.ExportResult{Data: []byte("ok"), ContentType: "text/plain", Filename: "test.txt"}, nil
}
