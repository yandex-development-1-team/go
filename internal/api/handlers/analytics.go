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

type analyticsExporter interface {
	Export(ctx context.Context, req dto.AnalyticsExportRequest) (apiService.ExportResult, error)
}

type AnalyticsHandler struct {
	svc analyticsExporter
}

func NewAnalyticsHandler(svc *apiService.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) Export(c *gin.Context) {
	exportType := dto.ExportType(c.Query("type"))
	switch exportType {
	case dto.ExportTypeBoxes, dto.ExportTypeUsers:
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
	default:
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверное значение format: допустимые значения — xlsx, csv"})
		return
	}

	dateFrom, err := parseOptionalDate(c.Query("date_from"))
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверный формат date_from: ожидается YYYY-MM-DD"})
		return
	}
	dateTo, err := parseOptionalDate(c.Query("date_to"))
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверный формат date_to: ожидается YYYY-MM-DD"})
		return
	}

	if dateFrom != nil && dateTo != nil && dateTo.Before(*dateFrom) {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"date_to не может быть раньше date_from"})
		return
	}

	result, err := h.svc.Export(c.Request.Context(), dto.AnalyticsExportRequest{
		Type:     exportType,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Format:   format,
	})
	if err != nil {
		apierrors.WriteErrorGin(c, err)
		return
	}

	c.Header("Content-Disposition", `attachment; filename="`+result.Filename+`"`)
	c.Data(http.StatusOK, result.ContentType, result.Data)
}

func parseOptionalDate(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(exportDateLayout, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
