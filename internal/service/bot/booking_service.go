package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yandex-development-1-team/go/internal/handlers/validation"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

// State constants of the booking process
const (
	StepStartBooking = iota
	StepSelectDate
	StepEnterName
	StepEnterOrg
	StepEnterPosition
	StepConfirmation
	StepMainMenu
	StepReturnInBoxList
)

// Constants
const (
	CallbackBookingPrefix = "book"
	KeyForBookingData     = "data"
)

// BookingState represents the booking status
type BookingState struct {
	UserID            int64
	ServiceID         int64
	SelectedSlot      models.BoxAvailableSlot
	GuestName         string
	GuestOrganization string
	GuestPosition     string
	Step              int
	CreatedAt         time.Time
}

// BookingService implements a booking service
type BookingService struct {
	session repository.SessionRepository
	repo    repository.BookingRepository
	boxRepo repository.BoxSolutionRepository
}

// NewBookingService creates a new instance of the booking service
func NewBookingService(
	session repository.SessionRepository,
	repo repository.BookingRepository,
	boxRepo repository.BoxSolutionRepository,
) *BookingService {
	return &BookingService{
		session: session,
		repo:    repo,
		boxRepo: boxRepo,
	}
}

// GetBookingState returns the state of the booking process
func (s *BookingService) GetBookingState(ctx context.Context, userID int64) *BookingState {
	session, err := s.session.GetSession(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil
	}

	stateData, ok := session.StateData[KeyForBookingData]
	if !ok {
		logger.Error("state data not found in session")
		return nil
	}

	jsonData, err := json.Marshal(stateData)
	if err != nil {
		logger.Error("failed to marshal state data", zap.Error(err))
		return nil
	}

	var state BookingState
	if err := json.Unmarshal(jsonData, &state); err != nil {
		logger.Error("failed to unmarshal state data", zap.Error(err))
		return nil
	}

	return &state
}

// SaveSession saves the booking state
func (s *BookingService) SaveSession(ctx context.Context, userID int64, stateData interface{}) error {
	data := map[string]interface{}{
		KeyForBookingData: stateData,
	}

	err := s.session.SaveSession(ctx, userID, CallbackBookingPrefix, data)
	if err != nil {
		logger.Error("failed to save user session",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return err
	}
	return nil
}

// CreateSession creates a session
func (s *BookingService) CreateSession(ctx context.Context, userID int64, serviceID int64) (*BookingState, error) {
	state := &BookingState{
		UserID:       userID,
		ServiceID:    serviceID,
		SelectedSlot: models.BoxAvailableSlot{},
		Step:         StepSelectDate,
		CreatedAt:    time.Now(),
	}

	err := s.SaveSession(ctx, userID, *state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// CreateBooking creates a new booking from state
func (s *BookingService) CreateBooking(ctx context.Context, state *BookingState) (int64, error) {
	date, err := time.Parse("2006-01-02", state.SelectedSlot.Date)
	if err != nil {
		return 0, err
	}

	booking := &models.Booking{
		UserID:            state.UserID,
		ServiceID:         int16(state.ServiceID),
		BookingDate:       date,
		BookingTime:       nil,
		GuestName:         state.GuestName,
		GuestOrganization: state.GuestOrganization,
		GuestPosition:     state.GuestPosition,
		Status:            "confirmation",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.ClearSession(ctx, state.UserID); err != nil {
		return 0, fmt.Errorf("clear session: %w", err)
	}
	return s.repo.CreateBooking(ctx, booking)
}

// ClearSession clears the user session
func (s *BookingService) ClearSession(ctx context.Context, userID int64) error {
	return s.session.ClearSession(ctx, userID)
}

// ValidateAndSetName validates and sets guest name
func (s *BookingService) ValidateAndSetName(ctx context.Context, state *BookingState, name string) error {
	if err := validation.Name(name); err != nil {
		return err
	}
	state.GuestName = name
	state.Step = StepEnterOrg
	err := s.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// ValidateAndSetOrganization validates and sets organization
func (s *BookingService) ValidateAndSetOrganization(ctx context.Context, state *BookingState, org string) error {
	if err := validation.Organization(org); err != nil {
		return err
	}
	state.GuestOrganization = org
	state.Step = StepEnterPosition
	err := s.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// ValidateAndSetPosition validates and sets position
func (s *BookingService) ValidateAndSetPosition(ctx context.Context, state *BookingState, position string) error {
	if err := validation.Position(position); err != nil {
		return err
	}
	state.GuestPosition = position
	state.Step = StepConfirmation
	err := s.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		return err
	}
	return nil
}

// ProcessDateSelection processes date selection and returns next step
func (s *BookingService) ProcessDateSelection(ctx context.Context, state *BookingState, slot models.BoxAvailableSlot) (bool, error) {
	_, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		return false, err
	}

	_, err = time.Parse("15:04", slot.StartTime)
	if err != nil {
		return false, err
	}

	_, err = time.Parse("15:04", slot.EndTime)
	if err != nil {
		return false, err
	}

	available, err := s.boxRepo.CheckSlotAvailability(ctx, state.ServiceID, slot)
	if err != nil {
		return false, err
	}

	if !available {
		state.Step = StepStartBooking
	} else {
		state.SelectedSlot = slot
		state.Step = StepEnterName
	}

	err = s.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		logger.Error("status save error", zap.Error(err))
		return false, err
	}

	return available, nil
}

// GetAvailableDates gets the available dates for the service
func (s *BookingService) GetAvailableSlots(ctx context.Context, serviceID int64) ([]models.BoxAvailableSlot, error) {
	slots, err := s.boxRepo.GetAvailableSlotsByServiceID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	return slots, nil
}

// IsBookingComplete checks if all required fields are filled
func (s *BookingService) IsBookingComplete(state *BookingState) bool {
	return state.GuestName != "" &&
		state.GuestOrganization != "" &&
		state.GuestPosition != "" &&
		state.SelectedSlot.Date != ""
}
