package inmemory

import (
	"fmt"
	"sync"
	"time"
)

// BookingState представляет состояние бронирования
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

// BookingStorage представляет потокобезопасное in-memory хранилище
type BookingStorage struct {
	mu     sync.RWMutex
	states map[int64]*BookingState
}

// NewBookingStorage создает новое хранилище
func NewBookingStorage() *BookingStorage {
	return &BookingStorage{
		states: make(map[int64]*BookingState),
	}
}

// Save сохраняет или обновляет состояние бронирования для пользователя
func (s *BookingStorage) Save(userID int64, state *BookingState) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stateCopy := *state
	s.states[userID] = &stateCopy

	return nil
}

// Get возвращает состояние бронирования для пользователя
func (s *BookingStorage) Get(userID int64) (*BookingState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, exists := s.states[userID]
	if !exists {
		return nil, fmt.Errorf("state not found for user %d", userID)
	}

	stateCopy := *state
	return &stateCopy, nil
}

// GetOrCreate возвращает существующее состояние или создает новое
func (s *BookingStorage) GetOrCreate(userID int64) *BookingState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if state, exists := s.states[userID]; exists {
		stateCopy := *state
		return &stateCopy
	}

	newState := &BookingState{
		UserID:    userID,
		Step:      1,
		CreatedAt: time.Now(),
	}

	s.states[userID] = newState

	newStateCopy := *newState
	return &newStateCopy
}

// Update обновляет определенные поля состояния
func (s *BookingStorage) Update(userID int64, updateFn func(*BookingState) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.states[userID]
	if !exists {
		return fmt.Errorf("state not found for user %d", userID)
	}

	if err := updateFn(state); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	return nil
}

// Delete удаляет состояние для пользователя
func (s *BookingStorage) Delete(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, userID)
}

// GetAll возвращает все состояния (для отладки)
func (s *BookingStorage) GetAll() map[int64]*BookingState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int64]*BookingState, len(s.states))
	for userID, state := range s.states {
		stateCopy := *state
		result[userID] = &stateCopy
	}

	return result
}

// Count возвращает количество состояний в хранилище
func (s *BookingStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.states)
}

// Clear очищает все хранилище
func (s *BookingStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states = make(map[int64]*BookingState)
}

// Cleanup удаляет старые записи
func (s *BookingStorage) Cleanup(olderThan time.Time) int {
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
