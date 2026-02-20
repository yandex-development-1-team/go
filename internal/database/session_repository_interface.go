package repository

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/models"
)

// SessionRepository defines the contract for session persistence.
type SessionRepository interface {
	// SaveSession creates or fully replaces the session for the given user.
	SaveSession(ctx context.Context, userID int64, state string, data map[string]interface{}) error

	// GetSession retrieves the session for the given user.
	// Returns ErrSessionNotFound if no session exists.
	GetSession(ctx context.Context, userID int64) (*models.UserSession, error)

	// ClearSession removes the session for the given user.
	// A no-op (no error) if the session does not exist.
	ClearSession(ctx context.Context, userID int64) error

	// UpdateSessionState changes only the state field of an existing session,
	// preserving all existing state_data.
	// Returns ErrSessionNotFound if no session exists.
	UpdateSessionState(ctx context.Context, userID int64, newState string) error
}
