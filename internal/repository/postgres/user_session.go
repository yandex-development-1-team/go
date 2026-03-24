package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	createsSessionQuery = `
		INSERT INTO user_sessions (user_id, current_state, state_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			current_state = EXCLUDED.current_state,
			state_data = EXCLUDED.state_data,
			updated_at = EXCLUDED.updated_at
	`
	getSessionByUserIDQuery = `
		SELECT user_id, current_state, state_data, created_at, updated_at
		FROM user_sessions 
		WHERE user_id = $1
	`
	updateSessionStateQuery = `
		UPDATE user_sessions 
		SET current_state = $2, updated_at = $3
		WHERE user_id = $1
	`
	deleteSesionQuery = `
		DELETE FROM user_sessions WHERE user_id = $1
	`
)

type pgSessionRepo struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *pgSessionRepo {
	return &pgSessionRepo{db: db}
}

func (r *pgSessionRepo) SaveSession(
	ctx context.Context,
	userID int64,
	state string,
	data map[string]interface{},
) error {
	now := time.Now().UTC()

	var stateDataJSON []byte
	if data != nil {
		var err error
		stateDataJSON, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal state data for user %d: %w", userID, err)
		}
	} else {
		stateDataJSON = []byte("null")
	}

	_, err := r.db.ExecContext(ctx, createsSessionQuery, userID, state, stateDataJSON, now, now)
	if err != nil {
		return fmt.Errorf("upsert session for user %d: %w", userID, err)
	}

	return nil
}

func (r *pgSessionRepo) GetSession(ctx context.Context, userID int64) (*models.UserSession, error) {
	var row struct {
		ID           int64     `db:"id"`
		UserID       int64     `db:"user_id"`
		CurrentState string    `db:"current_state"`
		StateData    []byte    `db:"state_data"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}

	err := r.db.GetContext(ctx, &row, getSessionByUserIDQuery, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session for user %d: %w", userID, err)
	}

	var stateData map[string]interface{}
	if len(row.StateData) > 0 && string(row.StateData) != "null" {
		if err := json.Unmarshal(row.StateData, &stateData); err != nil {
			return nil, fmt.Errorf("unmarshal state data for user %d: %w", userID, err)
		}
	} else {
		stateData = make(map[string]interface{})
	}

	return &models.UserSession{
		UserID:       row.UserID,
		CurrentState: row.CurrentState,
		StateData:    stateData,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func (r *pgSessionRepo) ClearSession(ctx context.Context, userID int64) error {
	result, err := r.db.ExecContext(ctx, deleteSesionQuery, userID)
	if err != nil {
		return fmt.Errorf("delete session for user %d: %w", userID, err)
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}

func (r *pgSessionRepo) UpdateSessionState(ctx context.Context, userID int64, newState string) error {
	result, err := r.db.ExecContext(ctx, updateSessionStateQuery, userID, newState, time.Now().UTC())

	if err != nil {
		return fmt.Errorf("update session state for user %d: %w", userID, err)
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}
