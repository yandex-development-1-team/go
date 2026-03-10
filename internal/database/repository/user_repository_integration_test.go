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

// Тесты используют общий TestMain и repoUser из booking_repository_test.go (контейнер Postgres + миграции).
// Запуск: go test -tags=integration ./internal/database/repository/... или make test-integration.

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		reqUserName     string
		reqFirstName    string
		reqLastName     string
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   999001,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "double_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   999001,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   999002,
			reqUserName:     "",
			reqFirstName:    "",
			reqLastName:     "",
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   999003,
			reqUserName:     "test_name",
			reqFirstName:    "test_first_name",
			reqLastName:     "test_last_name",
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			err := repoUser.CreateUser(ctx, tt.reqTelegramId, tt.reqUserName, tt.reqFirstName, tt.reqLastName)
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		wantUserName    string
		wantFirstName   string
		wantLastName    string
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantUserName:    "test_name",
			wantFirstName:   "test_first_name",
			wantLastName:    "test_last_name",
			wantErr:         nil,
		},
		{
			name:            "user_not_found",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantErr:         models.ErrUserNotFound,
		},
		{
			name:            "request_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "request_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   123456,
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "request_canceled" {
				cancel()
			}

			user, err := repoUser.GetUserByTelegramID(ctx, tt.reqTelegramId)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
				return
			}
			require.NoError(t, err, tt.name)
			require.NotNil(t, user)
			assert.Positive(t, user.ID)
			assert.Equal(t, tt.wantUserName, user.Username)
			assert.Equal(t, tt.wantFirstName, user.FirstName)
			assert.Equal(t, tt.wantLastName, user.LastName)
		})
	}
}

func TestUpdateUserGrade(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		reqGrade        int
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         nil,
		},
		{
			name:            "wrong_user",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			reqGrade:        2,
			wantErr:         models.ErrUserNotFound,
		},
		{
			name:            "context_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "context_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   123456,
			reqGrade:        2,
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "context_canceled" {
				cancel()
			}

			err := repoUser.UpdateUserGrade(ctx, tt.reqTelegramId, tt.reqGrade)
			assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name            string
		contextDuration time.Duration
		reqTelegramId   int64
		wantAdmin       bool
		wantErr         error
	}{
		{
			name:            "correct_data",
			contextDuration: 2 * time.Second,
			reqTelegramId:   123456,
			wantAdmin:       true,
			wantErr:         nil,
		},
		{
			name:            "wrong_user",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         models.ErrUserNotFound,
		},
		{
			name:            "context_canceled",
			contextDuration: 2 * time.Second,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         models.ErrRequestCanceled,
		},
		{
			name:            "context_timeout",
			contextDuration: 1 * time.Nanosecond,
			reqTelegramId:   222222,
			wantAdmin:       false,
			wantErr:         models.ErrRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.contextDuration)
			defer cancel()
			if tt.name == "context_canceled" {
				cancel()
			}

			isAdmin, err := repoUser.IsAdmin(ctx, tt.reqTelegramId)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "%s: actual: %v, expected: %v", tt.name, err, tt.wantErr)
				return
			}
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.wantAdmin, isAdmin, tt.name)
		})
	}
}
