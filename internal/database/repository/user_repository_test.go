package repository

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"github.com/yandex-development-1-team/go/internal/models"
// )

// func TestCreateUser(t *testing.T) {
// 	userRepo := NewUserRepository(db)

// 	tests := []struct {
// 		name          string
// 		timeout       time.Duration
// 		cancelBefore  bool
// 		telegramID    int64
// 		userName      string
// 		firstName     string
// 		lastName      string
// 		wantErr       error
// 	}{
// 		{
// 			name:       "correct_data",
// 			timeout:    2 * time.Second,
// 			telegramID: 123456,
// 			userName:   "test_name",
// 			firstName:  "test_first_name",
// 			lastName:   "test_last_name",
// 		},
// 		{
// 			name:       "upsert_same_user",
// 			timeout:    2 * time.Second,
// 			telegramID: 123456,
// 			userName:   "updated_name",
// 			firstName:  "test_first_name",
// 			lastName:   "test_last_name",
// 		},
// 		{
// 			name:         "request_canceled",
// 			timeout:      2 * time.Second,
// 			cancelBefore: true,
// 			telegramID:   123456,
// 			userName:     "test_name",
// 			firstName:    "test_first_name",
// 			lastName:     "test_last_name",
// 			wantErr:      models.ErrRequestCanceled,
// 		},
// 		{
// 			name:       "request_timeout",
// 			timeout:    1 * time.Nanosecond,
// 			telegramID: 123456,
// 			userName:   "test_name",
// 			firstName:  "test_first_name",
// 			lastName:   "test_last_name",
// 			wantErr:    models.ErrRequestTimeout,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
// 			defer cancel()

// 			if tt.cancelBefore {
// 				cancel()
// 			}

// 			err := userRepo.CreateUser(ctx, tt.telegramID, tt.userName, tt.firstName, tt.lastName)
// 			if tt.wantErr != nil {
// 				assert.ErrorIs(t, err, tt.wantErr)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestGetUser(t *testing.T) {
// 	userRepo := NewUserRepository(db)

// 	_, err := db.Exec(`
// 		INSERT INTO users (telegram_id, username, first_name, last_name)
// 		VALUES ($1, $2, $3, $4)
// 		ON CONFLICT (telegram_id) DO UPDATE SET username = EXCLUDED.username`,
// 		123456, "test_name", "test_first_name", "test_last_name")
// 	require.NoError(t, err)

// 	tests := []struct {
// 		name         string
// 		timeout      time.Duration
// 		cancelBefore bool
// 		telegramID   int64
// 		wantUserName string
// 		wantFirst    string
// 		wantLast     string
// 		wantErr      error
// 	}{
// 		{
// 			name:         "correct_data",
// 			timeout:      2 * time.Second,
// 			telegramID:   123456,
// 			wantUserName: "test_name",
// 			wantFirst:    "test_first_name",
// 			wantLast:     "test_last_name",
// 		},
// 		{
// 			name:       "user_not_found",
// 			timeout:    2 * time.Second,
// 			telegramID: 222222,
// 			wantErr:    models.ErrUserNotFound,
// 		},
// 		{
// 			name:         "request_canceled",
// 			timeout:      2 * time.Second,
// 			cancelBefore: true,
// 			telegramID:   123456,
// 			wantErr:      models.ErrRequestCanceled,
// 		},
// 		{
// 			name:       "request_timeout",
// 			timeout:    1 * time.Nanosecond,
// 			telegramID: 123456,
// 			wantErr:    models.ErrRequestTimeout,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
// 			defer cancel()

// 			if tt.cancelBefore {
// 				cancel()
// 			}

// 			user, err := userRepo.GetUserByTelegramID(ctx, tt.telegramID)
// 			if tt.wantErr != nil {
// 				assert.ErrorIs(t, err, tt.wantErr)
// 				return
// 			}

// 			require.NoError(t, err)
// 			assert.Equal(t, tt.wantUserName, user.Username)
// 			assert.Equal(t, tt.wantFirst, user.FirstName)
// 			assert.Equal(t, tt.wantLast, user.LastName)
// 		})
// 	}
// }

// func TestUpdateUserGrade(t *testing.T) {
// 	userRepo := NewUserRepository(db)

// 	tests := []struct {
// 		name         string
// 		timeout      time.Duration
// 		cancelBefore bool
// 		telegramID   int64
// 		grade        int
// 		wantErr      error
// 	}{
// 		{
// 			name:       "correct_data",
// 			timeout:    2 * time.Second,
// 			telegramID: 123456,
// 			grade:      2,
// 		},
// 		{
// 			name:       "user_not_found",
// 			timeout:    2 * time.Second,
// 			telegramID: 222222,
// 			grade:      2,
// 			wantErr:    models.ErrUserNotFound,
// 		},
// 		{
// 			name:         "context_canceled",
// 			timeout:      2 * time.Second,
// 			cancelBefore: true,
// 			telegramID:   123456,
// 			grade:        2,
// 			wantErr:      models.ErrRequestCanceled,
// 		},
// 		{
// 			name:       "context_timeout",
// 			timeout:    1 * time.Nanosecond,
// 			telegramID: 123456,
// 			grade:      2,
// 			wantErr:    models.ErrRequestTimeout,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
// 			defer cancel()

// 			if tt.cancelBefore {
// 				cancel()
// 			}

// 			err := userRepo.UpdateUserGrade(ctx, tt.telegramID, tt.grade)
// 			if tt.wantErr != nil {
// 				assert.ErrorIs(t, err, tt.wantErr)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestIsAdmin(t *testing.T) {
// 	userRepo := NewUserRepository(db)

// 	tests := []struct {
// 		name         string
// 		timeout      time.Duration
// 		cancelBefore bool
// 		telegramID   int64
// 		wantAdmin    bool
// 		wantErr      error
// 	}{
// 		{
// 			name:       "existing_user",
// 			timeout:    2 * time.Second,
// 			telegramID: 123456,
// 			wantAdmin:  false,
// 		},
// 		{
// 			name:       "user_not_found",
// 			timeout:    2 * time.Second,
// 			telegramID: 222222,
// 			wantAdmin:  false,
// 			wantErr:    models.ErrUserNotFound,
// 		},
// 		{
// 			name:         "context_canceled",
// 			timeout:      2 * time.Second,
// 			cancelBefore: true,
// 			telegramID:   222222,
// 			wantAdmin:    false,
// 			wantErr:      models.ErrRequestCanceled,
// 		},
// 		{
// 			name:       "context_timeout",
// 			timeout:    1 * time.Nanosecond,
// 			telegramID: 222222,
// 			wantAdmin:  false,
// 			wantErr:    models.ErrRequestTimeout,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
// 			defer cancel()

// 			if tt.cancelBefore {
// 				cancel()
// 			}

// 			isAdmin, err := userRepo.IsAdmin(ctx, tt.telegramID)
// 			if tt.wantErr != nil {
// 				assert.ErrorIs(t, err, tt.wantErr)
// 			} else {
// 				require.NoError(t, err)
// 				assert.Equal(t, tt.wantAdmin, isAdmin)
// 			}
// 		})
// 	}
// }
