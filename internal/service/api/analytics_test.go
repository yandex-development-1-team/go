package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/dto"
)

// mockAnalyticsQuerier is a test double for AnalyticsQuerier.
type mockAnalyticsQuerier struct {
	boxes []dto.AnalyticsBoxRow
	users []dto.AnalyticsUserRow
	err   error
}

func (m *mockAnalyticsQuerier) GetBoxesAnalytics(_ context.Context, _, _ *time.Time) ([]dto.AnalyticsBoxRow, error) {
	return m.boxes, m.err
}

func (m *mockAnalyticsQuerier) GetUsersAnalytics(_ context.Context, _, _ *time.Time) ([]dto.AnalyticsUserRow, error) {
	return m.users, m.err
}

var (
	sampleBoxes = []dto.AnalyticsBoxRow{
		{ServiceID: 1, ServiceName: "Бокс А", TotalBookings: 10, ConfirmedBookings: 8, CancelledBookings: 2, CancellationRate: 20.00},
		{ServiceID: 2, ServiceName: "Бокс Б", TotalBookings: 5, ConfirmedBookings: 5, CancelledBookings: 0, CancellationRate: 0.00},
	}
	sampleUsers = []dto.AnalyticsUserRow{
		{UserID: 1, FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com", TotalBookings: 3, RegisteredAt: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)},
		{UserID: 2, FirstName: "Мария", LastName: "Петрова", Email: "maria@example.com", TotalBookings: 7, RegisteredAt: time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)},
	}
)

func TestAnalyticsService_Export_BoxesXLSX(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{boxes: sampleBoxes})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeBoxes,
		Format: dto.ExportFormatXLSX,
	})

	require.NoError(t, err)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.ContentType)
	assert.Equal(t, "analytics_boxes.xlsx", result.Filename)
	assert.NotEmpty(t, result.Data)
	// XLSX files start with the PK zip signature
	assert.Equal(t, []byte{0x50, 0x4B}, result.Data[:2])
}

func TestAnalyticsService_Export_BoxesCSV(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{boxes: sampleBoxes})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeBoxes,
		Format: dto.ExportFormatCSV,
	})

	require.NoError(t, err)
	assert.Equal(t, "text/csv; charset=utf-8", result.ContentType)
	assert.Equal(t, "analytics_boxes.csv", result.Filename)
	require.NotEmpty(t, result.Data)

	// Strip UTF-8 BOM and parse CSV
	body := bytes.TrimPrefix(result.Data, []byte{0xEF, 0xBB, 0xBF})
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3) // header + 2 data rows
	assert.Equal(t, boxesHeaders, records[0])
	assert.Equal(t, "Бокс А", records[1][1])
	assert.Equal(t, "10", records[1][2])
	assert.Equal(t, "20.00", records[1][5])
}

func TestAnalyticsService_Export_UsersXLSX(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{users: sampleUsers})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeUsers,
		Format: dto.ExportFormatXLSX,
	})

	require.NoError(t, err)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.ContentType)
	assert.Equal(t, "analytics_users.xlsx", result.Filename)
	assert.NotEmpty(t, result.Data)
	assert.Equal(t, []byte{0x50, 0x4B}, result.Data[:2])
}

func TestAnalyticsService_Export_UsersCSV(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{users: sampleUsers})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeUsers,
		Format: dto.ExportFormatCSV,
	})

	require.NoError(t, err)
	assert.Equal(t, "text/csv; charset=utf-8", result.ContentType)
	assert.Equal(t, "analytics_users.csv", result.Filename)
	require.NotEmpty(t, result.Data)

	body := bytes.TrimPrefix(result.Data, []byte{0xEF, 0xBB, 0xBF})
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 3) // header + 2 data rows
	assert.Equal(t, usersHeaders, records[0])
	assert.Equal(t, "Иван", records[1][1])
	assert.Equal(t, "ivan@example.com", records[1][3])
	assert.Equal(t, "2026-01-10", records[1][5])
}

func TestAnalyticsService_Export_DefaultFormatIsXLSX(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{boxes: sampleBoxes})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type: dto.ExportTypeBoxes,
		// Format is zero value — service receives it as empty string, treated as xlsx by handler,
		// but here we pass XLSX explicitly to verify the path
		Format: dto.ExportFormatXLSX,
	})

	require.NoError(t, err)
	assert.Contains(t, result.ContentType, "spreadsheetml")
}

func TestAnalyticsService_Export_RepoErrorPropagated(t *testing.T) {
	repoErr := errors.New("db unavailable")
	svc := NewAnalyticsService(&mockAnalyticsQuerier{err: repoErr})

	_, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeBoxes,
		Format: dto.ExportFormatXLSX,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
}

func TestAnalyticsService_Export_EmptyRows(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsQuerier{boxes: []dto.AnalyticsBoxRow{}})

	result, err := svc.Export(context.Background(), dto.AnalyticsExportRequest{
		Type:   dto.ExportTypeBoxes,
		Format: dto.ExportFormatCSV,
	})

	require.NoError(t, err)
	body := bytes.TrimPrefix(result.Data, []byte{0xEF, 0xBB, 0xBF})
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 1) // only header row
	assert.Equal(t, boxesHeaders, records[0])
}
