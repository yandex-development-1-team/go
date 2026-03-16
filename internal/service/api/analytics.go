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

// AnalyticsQuerier is the repository interface required by AnalyticsService.
// Satisfied by *repository.AnalyticsRepo.
type AnalyticsQuerier interface {
	GetBoxesAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsBoxRow, error)
	GetUsersAnalytics(ctx context.Context, dateFrom, dateTo *time.Time) ([]dto.AnalyticsUserRow, error)
}

// ExportResult carries the generated file bytes and its HTTP response metadata.
type ExportResult struct {
	Data        []byte
	ContentType string
	Filename    string
}

// AnalyticsService builds export files from analytics data.
type AnalyticsService struct {
	repo AnalyticsQuerier
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(repo AnalyticsQuerier) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// Export fetches data and generates a file according to the request parameters.
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
		data, err := buildCSV(boxesHeaders, func(w *csv.Writer) error {
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
		if err != nil {
			return ExportResult{}, err
		}
		return ExportResult{Data: data, ContentType: "text/csv; charset=utf-8", Filename: "analytics_boxes.csv"}, nil
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Боксы"
	f.SetSheetName("Sheet1", sheet)
	writeExcelHeaders(f, sheet, boxesHeaders)
	for i, r := range rows {
		row := i + 2
		_ = f.SetCellInt(sheet, excelCell(1, row), r.ServiceID)
		_ = f.SetCellStr(sheet, excelCell(2, row), r.ServiceName)
		_ = f.SetCellInt(sheet, excelCell(3, row), r.TotalBookings)
		_ = f.SetCellInt(sheet, excelCell(4, row), r.ConfirmedBookings)
		_ = f.SetCellInt(sheet, excelCell(5, row), r.CancelledBookings)
		_ = f.SetCellFloat(sheet, excelCell(6, row), r.CancellationRate, 2, 64)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{
		Data:        buf.Bytes(),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "analytics_boxes.xlsx",
	}, nil
}

func buildUsersFile(rows []dto.AnalyticsUserRow, format dto.ExportFormat) (ExportResult, error) {
	if format == dto.ExportFormatCSV {
		data, err := buildCSV(usersHeaders, func(w *csv.Writer) error {
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
		if err != nil {
			return ExportResult{}, err
		}
		return ExportResult{Data: data, ContentType: "text/csv; charset=utf-8", Filename: "analytics_users.csv"}, nil
	}

	f := excelize.NewFile()
	defer f.Close()
	const sheet = "Пользователи"
	f.SetSheetName("Sheet1", sheet)
	writeExcelHeaders(f, sheet, usersHeaders)
	for i, r := range rows {
		row := i + 2
		_ = f.SetCellInt(sheet, excelCell(1, row), r.UserID)
		_ = f.SetCellStr(sheet, excelCell(2, row), r.FirstName)
		_ = f.SetCellStr(sheet, excelCell(3, row), r.LastName)
		_ = f.SetCellStr(sheet, excelCell(4, row), r.Email)
		_ = f.SetCellInt(sheet, excelCell(5, row), r.TotalBookings)
		_ = f.SetCellStr(sheet, excelCell(6, row), r.RegisteredAt.Format("2006-01-02"))
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{
		Data:        buf.Bytes(),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Filename:    "analytics_users.xlsx",
	}, nil
}

// buildCSV writes a UTF-8 BOM + header row + data rows into a buffer.
// The BOM ensures correct opening in Microsoft Excel.
func buildCSV(headers []string, fill func(*csv.Writer) error) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
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

// excelCell converts 1-based column and row indices to a cell address (e.g. 1,1 → "A1").
func excelCell(col, row int) string {
	colName, _ := excelize.ColumnNumberToName(col)
	return fmt.Sprintf("%s%d", colName, row)
}
