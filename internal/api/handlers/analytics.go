package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yandex-development-1-team/go/internal/apierrors"
	"github.com/yandex-development-1-team/go/internal/dto"
	apiService "github.com/yandex-development-1-team/go/internal/service/api"
)

const exportDateLayout = "2006-01-02"

// analyticsExporter is a local interface to allow mocking in tests.
type analyticsExporter interface {
	Export(ctx context.Context, req dto.AnalyticsExportRequest) (apiService.ExportResult, error)
}

// AnalyticsHandler handles analytics HTTP requests.
type AnalyticsHandler struct {
	svc analyticsExporter
}

// NewAnalyticsHandler creates a new AnalyticsHandler backed by the given service.
func NewAnalyticsHandler(svc *apiService.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// Export handles GET /api/v1/analytics/export
func (h *AnalyticsHandler) Export(c *gin.Context) {
	exportType := dto.ExportType(c.Query("type"))
	switch exportType {
	case dto.ExportTypeBoxes, dto.ExportTypeUsers:
		// valid
	case "":
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Параметр type обязателен (допустимые значения: boxes, users)"})
		return
	default:
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверное значение type: допустимые значения — boxes, users"})
		return
	}

	format := dto.ExportFormat(c.DefaultQuery("format", string(dto.ExportFormatXLSX)))
	switch format {
	case dto.ExportFormatXLSX, dto.ExportFormatCSV:
		// valid
	default:
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверное значение format: допустимые значения — xlsx, csv"})
		return
	}

	var dateFrom, dateTo *time.Time
	if v := c.Query("date_from"); v != "" {
		t, err := time.Parse(exportDateLayout, v)
		if err != nil {
			apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
				[]string{"Неверный формат date_from: ожидается YYYY-MM-DD"})
			return
		}
		dateFrom = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := time.Parse(exportDateLayout, v)
		if err != nil {
			apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
				[]string{"Неверный формат date_to: ожидается YYYY-MM-DD"})
			return
		}
		dateTo = &t
	}

	req := dto.AnalyticsExportRequest{
		Type:     exportType,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Format:   format,
	}

	result, err := h.svc.Export(c.Request.Context(), req)
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}
