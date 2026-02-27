package inmemory

import (
	"fmt"
	"sync"
	"time"
)

// BookingState represents the booking status
type BookingState struct {
	UserID            int64
	ServiceID         int
	VisitType         string
	SelectedDate      time.Time
	GuestName         string
	GuestOrganization string
	GuestPosition     string
	Step              int
	CreatedAt         time.Time
}

// BookingStateStorage represents thread-safe in-memory storage
type BookingStateStorage struct {
	mu     sync.RWMutex
	states map[int64]*BookingState
}

// NewBookingState creates a new storage
func NewBookingState() *BookingStateStorage {
	return &BookingStateStorage{
		states: make(map[int64]*BookingState),
	}
}

// Save saves or updates the booking status for the user
func (s *BookingStateStorage) Save(userID int64, state *BookingState) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stateCopy := *state
	s.states[userID] = &stateCopy

	return nil
}

// Get returns the booking status for the user
func (s *BookingStateStorage) Get(userID int64) (*BookingState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.states[userID]
	if !exists {
		return nil, fmt.Errorf("state not found for user %d", userID)
	}

	stateCopy := *state
	return &stateCopy, nil
}

// Delete deletes the status for the user
func (s *BookingStateStorage) Delete(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, userID)
}

// GetAll returns all states (for debugging)
func (s *BookingStateStorage) GetAll() map[int64]*BookingState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int64]*BookingState, len(s.states))
	for userID, state := range s.states {
		stateCopy := *state
		result[userID] = &stateCopy
	}

	return result
}

// Count returns the number of states in the storage
func (s *BookingStateStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.states)
}

// Clear cleans up the entire storage
func (s *BookingStateStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states = make(map[int64]*BookingState)
}

// Cleanup deletes old records
func (s *BookingStateStorage) Cleanup(olderThan time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for userID, state := range s.states {
		if state.CreatedAt.Before(olderThan) {
			delete(s.states, userID)
			removed++
		}
	}

	return removed
}
