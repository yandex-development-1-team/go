package handlers

import (
	"context"
	"errors"
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
	GetOverviewAnalytics(ctx context.Context, dateFrom *time.Time, dateTo *time.Time) (dto.AnalyticsOverview, error)
	GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error)
	GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error)
	GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error)
}

type AnalyticsHandler struct {
	svc analyticsExporter
}

func NewAnalyticsHandler(svc *apiService.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		svc: svc,
	}
}

func resolvePeriod(period string) (*time.Time, *time.Time) {
	now := time.Now()

	switch period {
	case "today":
		from := now.Truncate(24 * time.Hour)
		return &from, &now
	case "week":
		from := now.AddDate(0, 0, -7)
		return &from, &now
	case "month":
		from := now.AddDate(0, -1, 0)
		return &from, &now
	case "year":
		from := now.AddDate(-1, 0, 0)
		return &from, &now
	default:
		return nil, nil
	}
}

func (h *AnalyticsHandler) Export(c *gin.Context) {
	exportType := dto.ExportType(c.Query("type"))
	switch exportType {
	case dto.ExportTypeBoxes, dto.ExportTypeUsers:
	default:
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"тип должен быть boxes или users"})
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

func getDate(c *gin.Context) (*time.Time, *time.Time, error) {
	if period := c.Query("period"); period != "" {
		from, to := resolvePeriod(period)
		if from == nil || to == nil {
			return nil, nil, errors.New("invalid period")
		}
		return from, to, nil
	}

	dateFrom, err := parseOptionalDate(c.Query("date_from"))
	if err != nil {
		return nil, nil, err
	}

	dateTo, err := parseOptionalDate(c.Query("date_to"))
	if err != nil {
		return nil, nil, err
	}

	if dateFrom != nil && dateTo != nil && dateTo.Before(*dateFrom) {
		return nil, nil, errors.New("invalid date range")
	}

	return dateFrom, dateTo, nil
}

func (h *AnalyticsHandler) Overview(c *gin.Context) {
	dateFrom, dateTo, err := getDate(c)
	if err != nil {
		return
	}
	result, err := h.svc.GetOverviewAnalytics(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) GetBoxesAnalyticsExtended(c *gin.Context) {
	dateFrom, dateTo, err := getDate(c)
	if err != nil {
		return
	}

	sortBy := c.DefaultQuery("sort_by", "popularity")

	allowed := map[string]bool{
		"popularity": true,
		"revenue":    true,
		"attendance": true,
	}

	if !allowed[sortBy] {
		apierrors.WriteErrorMessagesGin(c, http.StatusBadRequest,
			[]string{"Неверное значение sort_by: допустимые — popularity, revenue, attendance"})
		return
	}

	result, err := h.svc.GetBoxesAnalyticsExtended(c.Request.Context(), dateFrom, dateTo, sortBy)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsBoxesResponse{
		Boxes: result,
	})
}

func (h *AnalyticsHandler) GetUsersAnalyticsExtended(c *gin.Context) {
	dateFrom, dateTo, err := getDate(c)
	if err != nil {
		return
	}

	result, err := h.svc.GetUsersAnalyticsExtended(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) GetDashboardAnalytics(c *gin.Context) {
	dateFrom, dateTo, err := getDate(c)
	if err != nil {
		return
	}

	result, err := h.svc.GetDashboardAnalytics(c.Request.Context(), dateFrom, dateTo)
	if err != nil {
		apierrors.WriteErrorMessagesGin(c, http.StatusInternalServerError,
			[]string{err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}