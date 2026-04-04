package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/yandex-development-1-team/go/internal/dto"
)

type AnalyticsQuerier interface {
	GetBoxesAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsBoxRow, error)
	GetUsersAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsUserRow, error)
	GetOverviewAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsOverview, error)
	GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error)
	GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error)
	GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error)
}

// ExportResult carries the generated file and its HTTP response metadata.
type ExportResult struct {
	Data        []byte
	ContentType string
	Filename    string
}

type AnalyticsService struct {
	repo AnalyticsQuerier
}

func NewAnalyticsService(repo AnalyticsQuerier) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) Export(ctx context.Context, req dto.AnalyticsExportRequest) (ExportResult, error) {
	switch req.Type {
	case dto.ExportTypeBoxes:
		rows, err := s.repo.GetBoxesAnalytics(ctx, req.DateFrom, req.DateTo)
		if err != nil {
			return ExportResult{}, err
		}
		return buildBoxesFile(rows, req.Format)
	case dto.ExportTypeUsers:
		rows, err := s.repo.GetUsersAnalytics(ctx, req.DateFrom, req.DateTo)
		if err != nil {
			return ExportResult{}, err
		}
		return buildUsersFile(rows, req.Format)
	default:
		return ExportResult{}, fmt.Errorf("unsupported export type: %s", req.Type)
	}
}

var boxesHeaders = []string{
	"ID сервиса", "Название", "Всего бронирований",
	"Подтверждённых", "Отменённых", "Процент отмен (%)",
}

var usersHeaders = []string{
	"ID пользователя", "Имя", "Фамилия", "Email",
	"Всего бронирований", "Дата регистрации",
}

func buildBoxesFile(rows []dto.AnalyticsBoxRow, format dto.ExportFormat) (ExportResult, error) {
	if format == dto.ExportFormatCSV {
		return csvResult("analytics_boxes.csv", boxesHeaders, func(w *csv.Writer) error {
			for _, r := range rows {
				if err := w.Write([]string{
					strconv.FormatInt(r.ServiceID, 10),
					r.ServiceName,
					strconv.FormatInt(r.TotalBookings, 10),
					strconv.FormatInt(r.ConfirmedBookings, 10),
					strconv.FormatInt(r.CancelledBookings, 10),
					strconv.FormatFloat(r.CancellationRate, 'f', 2, 64),
				}); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return xlsxResult("Боксы", "analytics_boxes.xlsx", boxesHeaders, func(f *excelize.File, sheet string) {
		for i, r := range rows {
			row := i + 2
			_ = f.SetCellInt(sheet, excelCell(1, row), r.ServiceID)
			_ = f.SetCellStr(sheet, excelCell(2, row), r.ServiceName)
			_ = f.SetCellInt(sheet, excelCell(3, row), r.TotalBookings)
			_ = f.SetCellInt(sheet, excelCell(4, row), r.ConfirmedBookings)
			_ = f.SetCellInt(sheet, excelCell(5, row), r.CancelledBookings)
			_ = f.SetCellFloat(sheet, excelCell(6, row), r.CancellationRate, 2, 64)
		}
	})
}

func buildUsersFile(rows []dto.AnalyticsUserRow, format dto.ExportFormat) (ExportResult, error) {
	if format == dto.ExportFormatCSV {
		return csvResult("analytics_users.csv", usersHeaders, func(w *csv.Writer) error {
			for _, r := range rows {
				if err := w.Write([]string{
					strconv.FormatInt(r.UserID, 10),
					r.FirstName,
					r.LastName,
					r.Email,
					strconv.FormatInt(r.TotalBookings, 10),
					r.RegisteredAt.Format("2006-01-02"),
				}); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return xlsxResult("Пользователи", "analytics_users.xlsx", usersHeaders, func(f *excelize.File, sheet string) {
		for i, r := range rows {
			row := i + 2
			_ = f.SetCellInt(sheet, excelCell(1, row), r.UserID)
			_ = f.SetCellStr(sheet, excelCell(2, row), r.FirstName)
			_ = f.SetCellStr(sheet, excelCell(3, row), r.LastName)
			_ = f.SetCellStr(sheet, excelCell(4, row), r.Email)
			_ = f.SetCellInt(sheet, excelCell(5, row), r.TotalBookings)
			_ = f.SetCellStr(sheet, excelCell(6, row), r.RegisteredAt.Format("2006-01-02"))
		}
	})
}

func csvResult(filename string, headers []string, fill func(*csv.Writer) error) (ExportResult, error) {
	data, err := buildCSV(headers, fill)
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Data: data, ContentType: "text/csv; charset=utf-8", Filename: filename}, nil
}

func xlsxResult(sheetName, filename string, headers []string, fillRows func(*excelize.File, string)) (ExportResult, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return ExportResult{}, err
	}
	writeExcelHeaders(f, sheetName, headers)
	fillRows(f, sheetName)
	buf, err := f.WriteToBuffer()
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{
		Data:        buf.Bytes(),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    filename,
	}, nil
}

func buildCSV(headers []string, fill func(*csv.Writer) error) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM for Excel compatibility
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	if err := fill(w); err != nil {
		return nil, err
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func writeExcelHeaders(f *excelize.File, sheet string, headers []string) {
	for i, h := range headers {
		_ = f.SetCellStr(sheet, excelCell(i+1, 1), h)
	}
}

// excelCell converts 1-based column and row to a cell address (e.g. 1,1 → "A1").
func excelCell(col, row int) string {
	colName, _ := excelize.ColumnNumberToName(col)
	return fmt.Sprintf("%s%d", colName, row)
}

func (s *AnalyticsService) GetOverviewAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsOverview, error) {
	return s.repo.GetOverviewAnalytics(ctx, dateFrom, dateTo)
}

func (s *AnalyticsService) GetBoxesAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time, sortBy string) ([]dto.AnalyticsBoxItem, error) {
	return s.repo.GetBoxesAnalyticsExtended(ctx, dateFrom, dateTo, sortBy)
}

func (s *AnalyticsService) GetUsersAnalyticsExtended(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsUsers, error) {
	return s.repo.GetUsersAnalyticsExtended(ctx, dateFrom, dateTo)
}

func (s *AnalyticsService) GetDashboardAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) (dto.AnalyticsDashboard, error) {
	return s.repo.GetDashboardAnalytics(ctx, dateFrom, dateTo)
}
