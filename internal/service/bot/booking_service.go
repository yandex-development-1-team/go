package bot

import (
	"context"
	"encoding/json"
	"time"

	"github.com/yandex-development-1-team/go/internal/database/repository"
	"github.com/yandex-development-1-team/go/internal/handlers/validation"
	"github.com/yandex-development-1-team/go/internal/logger"
	"github.com/yandex-development-1-team/go/internal/models"
	"go.uber.org/zap"
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
	StepReturnBack
)

// Constants
const (
	CallbackBookingPrefix = "book"
	KeyForBookingData     = "data"
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

// BookingService implements a booking service
type BookingService struct {
	session repository.SessionRepository
	repo    *repository.BookingRepo
	boxRepo *repository.BoxSolutionRepo
}

// NewBookingService creates a new instance of the booking service
func NewBookingService(
	session repository.SessionRepository,
	repo *repository.BookingRepo,
	boxRepo *repository.BoxSolutionRepo,
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
func (s *BookingService) CreateSession(ctx context.Context, userID int64, serviceID int, visitType string) (*BookingState, error) {
	state := &BookingState{
		UserID:    userID,
		ServiceID: serviceID,
		VisitType: visitType,
		Step:      StepSelectDate,
		CreatedAt: time.Now(),
	}

	err := s.SaveSession(ctx, userID, *state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// CreateBooking creates a new booking from state
func (s *BookingService) CreateBooking(ctx context.Context, state *BookingState) (int64, error) {
	booking := &models.Booking{
		UserID:            state.UserID,
		ServiceID:         int16(state.ServiceID),
		BookingDate:       state.SelectedDate,
		BookingTime:       nil,
		GuestName:         state.GuestName,
		GuestOrganization: state.GuestOrganization,
		GuestPosition:     state.GuestPosition,
		VisitType:         state.VisitType,
		Status:            "confirmation",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	s.ClearSession(ctx, state.UserID)
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
func (s *BookingService) ProcessDateSelection(ctx context.Context, state *BookingState, dateStr string) (bool, error) {
	selectedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false, err
	}

	timeSlots, err := s.boxRepo.GetAvailableTimeSlotsByDate(ctx, state.ServiceID, dateStr)
	if err != nil {
		return false, err
	}

	if len(timeSlots) == 0 {
		state.Step = StepStartBooking
	} else {
		state.SelectedDate = selectedDate
		state.Step = StepEnterName
	}

	err = s.SaveSession(ctx, state.UserID, *state)
	if err != nil {
		logger.Error("status save error", zap.Error(err))
		return false, err
	}

	return len(timeSlots) > 0, nil
}

// GetAvailableDates gets the available dates for the service
func (s *BookingService) GetAvailableDates(ctx context.Context, serviceID int, visitType string) ([]string, error) {
	slots, err := s.boxRepo.GetAvailableSlotsByServiceID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	dates := make([]string, 0, len(slots))
	for _, slot := range slots {
		dates = append(dates, slot.Date)
	}

	return dates, nil
}

// IsBookingComplete checks if all required fields are filled
func (s *BookingService) IsBookingComplete(state *BookingState) bool {
	return state.GuestName != "" &&
		state.GuestOrganization != "" &&
		state.GuestPosition != "" &&
		!state.SelectedDate.IsZero()
}
