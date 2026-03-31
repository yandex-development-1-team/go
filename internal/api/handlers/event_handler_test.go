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
		TRUNCATE TABLE users, services, boxes, events, bookings CASCADE;

		INSERT INTO users (id, telegram_id, email, password_hash, role, status) 
		VALUES (1, 123456789, 'test@example.com', 'hash', 'manager', 'active');

		INSERT INTO services (id, name, box_solution) 
		VALUES (1, 'Test Service', true);

		INSERT INTO boxes (id, title, status) 
		VALUES (1, 'Main Box', 'active'::box_status_type);
		
		INSERT INTO events (id, box_id, event_date, event_time, total_slots, occupied_slots, status)
		VALUES (100, 1, '2026-03-30', '12:00:00', 10, 1, 'active');

		INSERT INTO bookings (id, user_id, service_id, event_id, booking_date, guest_name, status)
		VALUES (1, 1, 1, 100, '2026-03-30', 'Tester John', 'confirmed');

		SELECT setval(pg_get_serial_sequence('users', 'id'), 1);
		SELECT setval(pg_get_serial_sequence('services', 'id'), 1);
		SELECT setval(pg_get_serial_sequence('boxes', 'id'), 1);
		SELECT setval(pg_get_serial_sequence('events', 'id'), 100);
		SELECT setval(pg_get_serial_sequence('bookings', 'id'), 1);
	`)
	require.NoError(t, err)

	server := setupEventServer(t, db)

	t.Run("Create and filter events", func(t *testing.T) {
		newEvents := []dto.EventCreateRequest{
			{BoxID: 1, Date: "2026-03-20", Time: "10:00:00", TotalSlots: 5},
			{BoxID: 1, Date: "2026-04-25", Time: "14:00:00", TotalSlots: 10},
		}

		for _, s := range newEvents {
			body, _ := json.Marshal(s)
			resp, err := http.Post(server.URL+"/api/v1/events", "application/json", bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			resp.Body.Close()
		}

		filterURL := fmt.Sprintf("%s/api/v1/events?box_id=1&date_from=2026-03-19&date_to=2026-03-21", server.URL)
		resp, err := http.Get(filterURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var listResp models.EventListResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))

		if len(listResp.Items) > 0 {
			actualDate := listResp.Items[0].Date.Format("2006-01-02")
			assert.Equal(t, "2026-03-20", actualDate)
		}
	})

	t.Run("Get detailed event with bookings", func(t *testing.T) {
		eventID := 100
		detailURL := fmt.Sprintf("%s/api/v1/events/%d", server.URL, eventID)

		resp, err := http.Get(detailURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var event models.EventWithBookings
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&event))

		assert.Equal(t, int64(eventID), event.ID)
		assert.Equal(t, 10, event.TotalSlots)
		require.NotEmpty(t, event.Bookings)
		assert.Equal(t, "Tester John", event.Bookings[0].GuestName)
	})
}
