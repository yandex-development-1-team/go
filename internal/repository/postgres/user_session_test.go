//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-development-1-team/go/internal/models"
)

// Тесты используют общий TestMain и repoSession из session_repository_test.go (контейнер Postgres + миграции).
// Запуск: go test -tags=integration ./internal/repository/postgres/... или make test-integration.

// createTestUser создает тестового пользователя
func createTestUser(t *testing.T, userID int64, telegramID int64) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO users (id, telegram_id, username, first_name, last_name, password_hash) 
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, userID, telegramID, "test_user_"+time.Now().Format("20060102150405"), "Test", "User", "hash")
	require.NoError(t, err)
}

// cleanupUserAndSession удаляет сессию и пользователя
func cleanupUserAndSession(t *testing.T, userID int64) {
	t.Helper()
	_, _ = db.Exec("DELETE FROM user_sessions WHERE user_id = $1", userID)
	_, _ = db.Exec("DELETE FROM users WHERE id = $1", userID)
}

func TestSaveSession(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		userID          int64
		telegramID      int64
		state           string
		data            map[string]interface{}
		wantErr         error
	}{
		{
			name:            "correct_data_insert",
			contextDuration: 2 * time.Second,
			userID:          888001,
			telegramID:      888001,
			state:           "main_menu",
			data:            map[string]interface{}{"step": 1, "form_data": "test"},
			wantErr:         nil,
		},
		{
			name:            "correct_data_update",
			contextDuration: 2 * time.Second,
			userID:          888001,
			telegramID:      888001,
			state:           "booking_form",
			data:            map[string]interface{}{"step": 2, "form_data": "updated"},
			wantErr:         nil,
		},
		{
			name:            "empty_data",
			contextDuration: 2 * time.Second,
			userID:          888002,
			telegramID:      888002,
			state:           "service_detail",
			data:            nil,
			wantErr:         nil,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			userID:          888003,
			telegramID:      888003,
			state:           "main_menu",
			data:            map[string]interface{}{"step": 1},
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			userID:          888004,
			telegramID:      888004,
			state:           "main_menu",
			data:            map[string]interface{}{"step": 1},
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем пользователя для теста
			createTestUser(t, tt.userID, tt.telegramID)

			// Очищаем данные после теста
			t.Cleanup(func() {
				cleanupUserAndSession(t, tt.userID)
			})

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			err := repoSession.SaveSession(ctx, tt.userID, tt.state, tt.data)
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		})
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		userID          int64
		telegramID      int64
		initialState    string
		initialData     map[string]interface{}
		wantState       string
		wantData        map[string]interface{}
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			userID:          887001,
			telegramID:      887001,
			initialState:    "test_state",
			initialData:     map[string]interface{}{"key": "value", "num": 42},
			wantState:       "test_state",
			wantData:        map[string]interface{}{"key": "value", "num": float64(42)},
			wantErr:         nil,
		},
		{
			name:            "session_not_found",
			contextDuration: 2 * time.Second,
			userID:          777777,
			telegramID:      777777,
			wantErr:         models.ErrSessionNotFound,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			userID:          887002,
			telegramID:      887002,
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			userID:          887003,
			telegramID:      887003,
			wantErr:         models.ErrRequestTimeout,
		},
		{
			name:            "empty_state_data",
			contextDuration: 2 * time.Second,
			userID:          887004,
			telegramID:      887004,
			initialState:    "",
			initialData:     nil,
			wantState:       "",
			wantData:        map[string]interface{}{},
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем пользователя для теста
			createTestUser(t, tt.userID, tt.telegramID)

			// Очищаем данные после теста
			t.Cleanup(func() {
				cleanupUserAndSession(t, tt.userID)
			})

			// Создаем сессию для тестов, которым она нужна
			if tt.name == "correct_data" || tt.name == "empty_state_data" {
				err := repoSession.SaveSession(context.Background(), tt.userID, tt.initialState, tt.initialData)
				require.NoError(t, err, "setup: save initial session failed")
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			session, err := repoSession.GetSession(ctx, tt.userID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
				return
			}
			require.NoError(t, err, tt.name)
			require.NotNil(t, session)
			assert.Equal(t, tt.userID, session.UserID)
			assert.Equal(t, tt.wantState, session.CurrentState)

			assert.Equal(t, tt.wantData, session.StateData)
			assert.NotZero(t, session.CreatedAt)
			assert.NotZero(t, session.UpdatedAt)
		})
	}
}

func TestClearSession(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		userID          int64
		telegramID      int64
		initialState    string
		initialData     map[string]interface{}
		wantErr         error
	}{
		{
			name:            "correct_delete",
			contextDuration: 2 * time.Second,
			userID:          886001,
			telegramID:      886001,
			initialState:    "test_state",
			initialData:     map[string]interface{}{"key": "value"},
			wantErr:         nil,
		},
		{
			name:            "session_not_found_after_delete",
			contextDuration: 2 * time.Second,
			userID:          886001,
			telegramID:      886001,
			initialState:    "test_state",
			initialData:     map[string]interface{}{"key": "value"},
			wantErr:         models.ErrSessionNotFound,
		},
		{
			name:            "non_existent_session",
			contextDuration: 2 * time.Second,
			userID:          666666,
			telegramID:      666666,
			wantErr:         models.ErrSessionNotFound,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			userID:          886002,
			telegramID:      886002,
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			userID:          886003,
			telegramID:      886003,
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем пользователя для теста
			createTestUser(t, tt.userID, tt.telegramID)

			// Очищаем данные после теста
			t.Cleanup(func() {
				cleanupUserAndSession(t, tt.userID)
			})

			// Создаем сессию для тестов, которым она нужна (кроме non_existent_session)
			if tt.name != "non_existent_session" {
				err := repoSession.SaveSession(context.Background(), tt.userID, tt.initialState, tt.initialData)
				if tt.name != "session_not_found_after_delete" {
					require.NoError(t, err, "setup: save initial session failed")
				}
			}

			// Для теста "session_not_found_after_delete" удаляем сессию вручную перед тем как пытаться удалить снова
			if tt.name == "session_not_found_after_delete" {
				err := repoSession.ClearSession(context.Background(), tt.userID)
				require.NoError(t, err, "setup: delete session for second attempt")
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			err := repoSession.ClearSession(ctx, tt.userID)
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		})
	}
}

func TestUpdateSessionState(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		userID          int64
		telegramID      int64
		initialState    string
		initialData     map[string]interface{}
		newState        string
		wantErr         error
	}{
		{
			name:            "correct_update",
			contextDuration: 2 * time.Second,
			userID:          885001,
			telegramID:      885001,
			initialState:    "initial_state",
			initialData:     map[string]interface{}{"key": "value"},
			newState:        "updated_state",
			wantErr:         nil,
		},
		{
			name:            "update_non_existent",
			contextDuration: 2 * time.Second,
			userID:          555555,
			telegramID:      555555,
			newState:        "another_state",
			wantErr:         models.ErrSessionNotFound,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			userID:          885002,
			telegramID:      885002,
			newState:        "test_state",
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			userID:          885003,
			telegramID:      885003,
			newState:        "test_state",
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// теста
			createTestUser(t, tt.userID, tt.telegramID)

			// Очищаем данные после теста
			t.Cleanup(func() {
				cleanupUserAndSession(t, tt.userID)
			})

			// Создаем сессию для теста "correct_update"
			if tt.name == "correct_update" {
				err := repoSession.SaveSession(context.Background(), tt.userID, tt.initialState, tt.initialData)
				require.NoError(t, err, "setup: save initial session failed")
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			err := repoSession.UpdateSessionState(ctx, tt.userID, tt.newState)
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		})
	}
}
