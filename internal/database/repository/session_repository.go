package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yandex-development-1-team/go/internal/config"
	"github.com/yandex-development-1-team/go/internal/models"
)

const (
	// defaultSessionTTL is the time-to-live for a session key when no custom TTL is set.
	defaultSessionTTL = 24 * time.Hour

	// keyPrefix namespaces all session keys in Redis.
	keyPrefix = "session:user:"
)

var (
	// ErrSessionNotFound is returned when a requested session does not exist.
	ErrSessionNotFound = errors.New("session not found")
)

type sessionDTO struct {
	UserID       int64                  `json:"user_id"`
	CurrentState string                 `json:"current_state"`
	StateData    map[string]interface{} `json:"state_data"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// SessionRepository is a Redis-backed implementation of repository.SessionRepository.
type SessionRepository struct {
	client redis.Cmdable
	ttl    time.Duration
}

// Option is a functional option for SessionRepository.
type Option func(*SessionRepository)

// WithTTL overrides the default session TTL.
func WithTTL(ttl time.Duration) Option {
	return func(r *SessionRepository) {
		r.ttl = ttl
	}
}

// NewSessionRepository constructs a new Redis-backed SessionRepository.
func NewSessionRepository(client redis.Cmdable, opts ...Option) *SessionRepository {
	r := &SessionRepository{
		client: client,
		ttl:    defaultSessionTTL,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// NewClient builds a *redis.Client from the supplied Config.
// Call client.Ping to verify connectivity before use.
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis: address must not be empty")
	}

	poolSize := cfg.PoolSize
	if poolSize == 0 {
		poolSize = 10
	}

	opts := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return redis.NewClient(opts), nil
}

// SaveSession creates or fully replaces the session for the given user.
// If a session already exists its CreatedAt timestamp is preserved.
func (r *SessionRepository) SaveSession(
	ctx context.Context,
	userID int64,
	state string,
	data map[string]interface{},
) error {
	now := time.Now().UTC()

	// Preserve CreatedAt when overwriting an existing session.
	createdAt := now
	if existing, err := r.getDTO(ctx, userID); err == nil {
		createdAt = existing.CreatedAt
	}

	dto := sessionDTO{
		UserID:       userID,
		CurrentState: state,
		StateData:    data,
		CreatedAt:    createdAt,
		UpdatedAt:    now,
	}

	return r.setDTO(ctx, userID, dto)
}

// GetSession retrieves the session for the given user.
// Returns repository.ErrSessionNotFound if no session exists.
func (r *SessionRepository) GetSession(ctx context.Context, userID int64) (*models.UserSession, error) {
	dto, err := r.getDTO(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dtoToModel(dto), nil
}

// ClearSession removes the session for the given user.
// It is a no-op (returns nil) when the session does not exist.
func (r *SessionRepository) ClearSession(ctx context.Context, userID int64) error {
	return withMetricsRedis("Del", func() error {

		key := r.buildKey(userID)
		if err := r.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("redis DEL %s: %w", key, err)
		}
		return nil
	})
}

// UpdateSessionState changes only the state field of an existing session,
// leaving all state_data intact.
// Returns repository.ErrSessionNotFound if the session does not exist.
func (r *SessionRepository) UpdateSessionState(ctx context.Context, userID int64, newState string) error {
	dto, err := r.getDTO(ctx, userID)
	if err != nil {
		return err
	}

	dto.CurrentState = newState
	dto.UpdatedAt = time.Now().UTC()

	return r.setDTO(ctx, userID, *dto)
}

func (r *SessionRepository) buildKey(userID int64) string {
	return fmt.Sprintf("%s%d", keyPrefix, userID)
}

// getDTO fetches and deserialises a session DTO from Redis.
func (r *SessionRepository) getDTO(ctx context.Context, userID int64) (*sessionDTO, error) {
	return withMetricsRedisValue("getDTO", func() (*sessionDTO, error) {

		key := r.buildKey(userID)

		raw, err := r.client.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, ErrSessionNotFound
			}
			return nil, fmt.Errorf("redis GET %s: %w", key, err)
		}

		var dto sessionDTO
		if err := json.Unmarshal(raw, &dto); err != nil {
			return nil, fmt.Errorf("unmarshal session %d: %w", userID, err)
		}
		return &dto, nil
	})
}

// setDTO serialises and stores a session DTO in Redis, refreshing the TTL.
func (r *SessionRepository) setDTO(ctx context.Context, userID int64, dto sessionDTO) error {
	return withMetricsRedis("setDTO", func() error {
		key := r.buildKey(userID)

		raw, err := json.Marshal(dto)
		if err != nil {
			return fmt.Errorf("marshal session %d: %w", userID, err)
		}

		if err := r.client.Set(ctx, key, raw, r.ttl).Err(); err != nil {
			return fmt.Errorf("redis SET %s: %w", key, err)
		}
		return nil
	})
}

// dtoToModel converts the internal DTO to the public domain model.
func dtoToModel(dto *sessionDTO) *models.UserSession {
	stateData := dto.StateData
	if stateData == nil {
		stateData = make(map[string]interface{})
	}
	return &models.UserSession{
		UserID:       dto.UserID,
		CurrentState: dto.CurrentState,
		StateData:    stateData,
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}
}
