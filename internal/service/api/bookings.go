package service

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type BookingsService struct {
	repo   repository.BookingRepository
	txRepo repository.TxRepository
}

func NewBookingsService(repo repository.BookingRepository, txRepo repository.TxRepository) *BookingsService {
	return &BookingsService{
		repo:   repo,
		txRepo: txRepo,
	}
}

func (s *BookingsService) GetBookingById(ctx context.Context, id int64) (*models.BookingAPI, error) {
	return s.repo.GetBookingById(ctx, id)
}

func (s *BookingsService) GetBookingsList(ctx context.Context, app *models.ApplicationFilter) (*models.BookingList, error) {
	return s.repo.GetBookingsList(ctx, app)
}

func (s *BookingsService) UpdateBookingStatus(ctx context.Context, id int64, status string) (*models.BookingAPI, error) {
	tx, err := s.txRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	txCtx := ctxutil.WithTx(ctx, tx)

	if err = s.repo.UpdateBookingStatus(txCtx, id, status); err != nil {
		return nil, err
	}

	var app *models.BookingAPI
	app, err = s.repo.GetBookingById(txCtx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *BookingsService) DeleteBooking(ctx context.Context, id int64) error {
	return s.repo.DeleteBooking(ctx, id)
}
