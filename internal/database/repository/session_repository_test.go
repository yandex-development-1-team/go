package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRepo spins up an isolated miniredis instance and returns a ready-to-use repository.
// Each test gets a clean instance — no shared state between tests.
func newTestRepo(t *testing.T, opts ...Option) (*SessionRepository, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t) // t.Cleanup is called automatically

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return NewSessionRepository(client, opts...), mr
}

// sessionKey returns the Redis key for a given userID, mirroring
// the buildKey logic from the repository to allow direct miniredis inspection.
func sessionKey(userID int64) string {
	return fmt.Sprintf("session:user:%d", userID)
}

// seedSession writes a session directly into miniredis, bypassing the repository.
// Used in tests that need to control the initial state independently of SaveSession.
func seedSession(t *testing.T, mr *miniredis.Miniredis, userID int64, state string, data map[string]interface{}, createdAt time.Time) {
	t.Helper()

	type dto struct {
		UserID       int64                  `json:"user_id"`
		CurrentState string                 `json:"current_state"`
		StateData    map[string]interface{} `json:"state_data"`
		CreatedAt    time.Time              `json:"created_at"`
		UpdatedAt    time.Time              `json:"updated_at"`
	}

	raw, err := json.Marshal(dto{
		UserID:       userID,
		CurrentState: state,
		StateData:    data,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	})
	require.NoError(t, err)
	require.NoError(t, mr.Set(sessionKey(userID), string(raw)))
}

// ─── SaveSession ─────────────────────────────────────────────────────────────

func TestSaveSession_NewSession_KeyAppearsInRedis(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t, WithTTL(time.Hour))
	ctx := context.Background()

	err := repo.SaveSession(ctx, 1, "main_menu", map[string]interface{}{"lang": "ru"})

	require.NoError(t, err)
	assert.True(t, mr.Exists(sessionKey(1)))
}

func TestSaveSession_NewSession_TTLIsApplied(t *testing.T) {
	t.Parallel()

	const ttl = 30 * time.Minute
	repo, mr := newTestRepo(t, WithTTL(ttl))
	ctx := context.Background()

	require.NoError(t, repo.SaveSession(ctx, 2, "main_menu", nil))

	remaining := mr.TTL(sessionKey(2))
	assert.InDelta(t, ttl.Seconds(), remaining.Seconds(), 1)
}

func TestSaveSession_NewSession_FieldsPersistedCorrectly(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	data := map[string]interface{}{"service_id": float64(5), "note": "vip"}
	require.NoError(t, repo.SaveSession(ctx, 3, "booking_form", data))

	session, err := repo.GetSession(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(3), session.UserID)
	assert.Equal(t, "booking_form", session.CurrentState)
	assert.Equal(t, float64(5), session.StateData["service_id"])
	assert.Equal(t, "vip", session.StateData["note"])
}

func TestSaveSession_Overwrite_PreservesCreatedAt(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(4)
	require.NoError(t, repo.SaveSession(ctx, userID, "main_menu", nil))

	first, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)
	originalCreatedAt := first.CreatedAt

	time.Sleep(5 * time.Millisecond)

	require.NoError(t, repo.SaveSession(ctx, userID, "booking_form", map[string]interface{}{"step": 1}))

	second, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)

	assert.True(t, second.CreatedAt.Equal(originalCreatedAt),
		"CreatedAt must be preserved when overwriting an existing session")
	assert.True(t, second.UpdatedAt.After(originalCreatedAt),
		"UpdatedAt must be refreshed on overwrite")
}

func TestSaveSession_Overwrite_ReplacesStateAndData(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(5)
	require.NoError(t, repo.SaveSession(ctx, userID, "main_menu", map[string]interface{}{"old_key": true}))
	require.NoError(t, repo.SaveSession(ctx, userID, "booking_form", map[string]interface{}{"new_key": true}))

	session, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)

	assert.Equal(t, "booking_form", session.CurrentState)
	assert.Equal(t, true, session.StateData["new_key"])
	assert.Nil(t, session.StateData["old_key"], "stale data must not survive a full session overwrite")
}

// ─── GetSession ───────────────────────────────────────────────────────────────

func TestGetSession_Exists_ReturnsAllFields(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(10)
	createdAt := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	seedSession(t, mr, userID, "service_detail", map[string]interface{}{
		"service_id": float64(42),
		"date":       "2024-12-01",
	}, createdAt)

	session, err := repo.GetSession(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, "service_detail", session.CurrentState)
	assert.Equal(t, float64(42), session.StateData["service_id"])
	assert.Equal(t, "2024-12-01", session.StateData["date"])
	assert.True(t, session.CreatedAt.Equal(createdAt))
}

func TestGetSession_NotFound_ReturnsErrSessionNotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	session, err := repo.GetSession(ctx, 9999)

	assert.Nil(t, session)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestGetSession_NilStateData_ReturnsEmptyMapNotNil(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	// Explicitly save nil data to verify the repository normalises it.
	require.NoError(t, repo.SaveSession(ctx, 11, "main_menu", nil))

	session, err := repo.GetSession(ctx, 11)

	require.NoError(t, err)
	assert.NotNil(t, session.StateData, "StateData must never be nil — callers rely on map access without nil checks")
	assert.Empty(t, session.StateData)
}

func TestGetSession_ComplexNestedData_RoundTrip(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	data := map[string]interface{}{
		"nested": map[string]interface{}{"key": "value"},
		"list":   []interface{}{"a", "b", "c"},
		"number": float64(99),
		"flag":   true,
	}
	require.NoError(t, repo.SaveSession(ctx, 12, "confirm_booking", data))

	session, err := repo.GetSession(ctx, 12)

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key": "value"}, session.StateData["nested"])
	assert.Equal(t, []interface{}{"a", "b", "c"}, session.StateData["list"])
	assert.Equal(t, float64(99), session.StateData["number"])
	assert.Equal(t, true, session.StateData["flag"])
}

func TestGetSession_AfterTTLExpiry_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t, WithTTL(time.Second))
	ctx := context.Background()

	require.NoError(t, repo.SaveSession(ctx, 13, "main_menu", nil))

	// FastForward advances miniredis internal clock without a real time.Sleep.
	mr.FastForward(2 * time.Second)

	_, err := repo.GetSession(ctx, 13)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

// ─── ClearSession ─────────────────────────────────────────────────────────────

func TestClearSession_Exists_KeyRemovedFromRedis(t *testing.T) {
	t.Parallel()

	repo, mr := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(20)
	require.NoError(t, repo.SaveSession(ctx, userID, "main_menu", nil))
	require.True(t, mr.Exists(sessionKey(userID)))

	err := repo.ClearSession(ctx, userID)

	require.NoError(t, err)
	assert.False(t, mr.Exists(sessionKey(userID)))
}

func TestClearSession_AfterClear_GetReturnsNotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(21)
	require.NoError(t, repo.SaveSession(ctx, userID, "main_menu", nil))
	require.NoError(t, repo.ClearSession(ctx, userID))

	_, err := repo.GetSession(ctx, userID)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}

func TestClearSession_NonExistent_NoError(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	// DEL on a missing key must be a no-op, not an error.
	err := repo.ClearSession(ctx, 8888)
	assert.NoError(t, err)
}

// ─── UpdateSessionState ───────────────────────────────────────────────────────

func TestUpdateSessionState_Exists_StateChanges(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(30)
	require.NoError(t, repo.SaveSession(ctx, userID, "booking_form", nil))

	err := repo.UpdateSessionState(ctx, userID, "confirm_booking")

	require.NoError(t, err)
	session, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "confirm_booking", session.CurrentState)
}

func TestUpdateSessionState_Exists_StateDataNotTouched(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(31)
	originalData := map[string]interface{}{"step": float64(2), "service": "massage"}
	require.NoError(t, repo.SaveSession(ctx, userID, "booking_form", originalData))

	require.NoError(t, repo.UpdateSessionState(ctx, userID, "payment_form"))

	session, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, "payment_form", session.CurrentState)
	assert.Equal(t, float64(2), session.StateData["step"],
		"state_data must not be modified by UpdateSessionState")
	assert.Equal(t, "massage", session.StateData["service"])
}

func TestUpdateSessionState_Exists_UpdatesUpdatedAt(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	const userID = int64(32)
	require.NoError(t, repo.SaveSession(ctx, userID, "main_menu", nil))

	before, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	require.NoError(t, repo.UpdateSessionState(ctx, userID, "booking_form"))

	after, err := repo.GetSession(ctx, userID)
	require.NoError(t, err)

	assert.True(t, after.UpdatedAt.After(before.UpdatedAt),
		"UpdatedAt must be refreshed after a state transition")
	assert.True(t, after.CreatedAt.Equal(before.CreatedAt),
		"CreatedAt must not change on state transition")
}

func TestUpdateSessionState_NotFound_ReturnsErrSessionNotFound(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	err := repo.UpdateSessionState(ctx, 7777, "any_state")

	assert.ErrorIs(t, err, ErrSessionNotFound)
}

// ─── Isolation ────────────────────────────────────────────────────────────────

func TestSessionIsolation_DifferentUsers_IndependentSessions(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SaveSession(ctx, 100, "main_menu", map[string]interface{}{"user": "alice"}))
	require.NoError(t, repo.SaveSession(ctx, 101, "booking_form", map[string]interface{}{"user": "bob"}))

	alice, err := repo.GetSession(ctx, 100)
	require.NoError(t, err)

	bob, err := repo.GetSession(ctx, 101)
	require.NoError(t, err)

	assert.Equal(t, "main_menu", alice.CurrentState)
	assert.Equal(t, "booking_form", bob.CurrentState)
	assert.Equal(t, "alice", alice.StateData["user"])
	assert.Equal(t, "bob", bob.StateData["user"])
}

func TestSessionIsolation_ClearOneUser_OtherUnaffected(t *testing.T) {
	t.Parallel()

	repo, _ := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SaveSession(ctx, 200, "main_menu", nil))
	require.NoError(t, repo.SaveSession(ctx, 201, "main_menu", nil))

	require.NoError(t, repo.ClearSession(ctx, 200))

	_, err := repo.GetSession(ctx, 200)
	assert.ErrorIs(t, err, ErrSessionNotFound, "session 200 must be deleted")

	_, err = repo.GetSession(ctx, 201)
	assert.NoError(t, err, "session 201 must not be affected")
}
