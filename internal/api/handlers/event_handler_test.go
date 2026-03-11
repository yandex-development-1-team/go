//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	repository "github.com/yandex-development-1-team/go/internal/repository/postgres"
)

func setupEventServer(t *testing.T, db *sqlx.DB) *httptest.Server {
	t.Helper()
	repo := repository.NewEventRepository(db)
	handler := NewEventHandler(repo)

	router := gin.New()
	router.POST("/api/v1/events", handler.CreateEvent)
	router.GET("/api/v1/events", handler.ListEvents)
	router.GET("/api/v1/events/:id", handler.GetEventByID)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func TestEventIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	container, err := startContainer()
	require.NoError(t, err)
	defer func() { _ = container.Terminate(ctx) }()

	db, err := createDB(container)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`
		-- Подготовка в services (на случай если FK ссылается сюда)
		INSERT INTO services (id, name, box_solution) 
		VALUES (1, 'Test Service', true) 
		ON CONFLICT (id) DO NOTHING;

		-- Подготовка в boxes (на случай если FK ссылается сюда)
		INSERT INTO boxes (id, title, status) 
		VALUES (1, 'Main Box', 'active'::box_status_type) 
		ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title;
		
		-- Синхронизация sequence
		SELECT setval(pg_get_serial_sequence('boxes', 'id'), (SELECT MAX(id) FROM boxes)) 
		WHERE EXISTS (SELECT 1 FROM pg_class WHERE relname = 'boxes');
	`)
	require.NoError(t, err)

	t.Logf("[DEBUG] Подготовка данных (ID=1) завершена успешно")

	server := setupEventServer(t, db)

	t.Run("Create and filter events", func(t *testing.T) {
		slots := []dto.EventCreateRequest{
			{BoxID: 1, Date: "2026-03-20", Time: "10:00:00", TotalSlots: 5},
			{BoxID: 1, Date: "2026-03-25", Time: "14:00:00", TotalSlots: 10},
		}

		for _, s := range slots {
			body, _ := json.Marshal(s)
			resp, err := http.Post(server.URL+"/api/v1/events", "application/json", bytes.NewReader(body))
			require.NoError(t, err)

			assert.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался 201 статус при создании события")
			resp.Body.Close()
		}

		filterURL := fmt.Sprintf("%s/api/v1/events?box_id=1&date_from=2026-03-19&date_to=2026-03-21", server.URL)
		resp, err := http.Get(filterURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var listResp dto.EventsListResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
		assert.Equal(t, 1, len(listResp.Items), "Фильтр по дате должен был вернуть ровно 1 элемент")
	})

	t.Run("Get detailed event", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/events?box_id=1")
		require.NoError(t, err)

		var listResp dto.EventsListResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
		resp.Body.Close()

		require.NotEmpty(t, listResp.Items, "Список ивентов пуст")

		eventID := listResp.Items[0].ID

		detailURL := fmt.Sprintf("%s/api/v1/events/%d", server.URL, eventID)
		resp, err = http.Get(detailURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var event models.EventWithBookings
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&event))

		assert.Equal(t, eventID, event.ID)
		assert.NotNil(t, event.Bookings, "Bookings должен быть пустым слайсом [], а не nil")
	})
}
