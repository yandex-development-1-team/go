package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type ApplicationsService struct {
	repo   repository.ApplicationRepository
	txRepo repository.TxRepository
}

func NewApplicationsService(repo repository.ApplicationRepository, txRepo repository.TxRepository) *ApplicationsService {
	return &ApplicationsService{
		repo:   repo,
		txRepo: txRepo,
	}
}

func (s *ApplicationsService) Create(ctx context.Context, req *models.Application) error {
	normalized_nickname, err := normalizeTelegramHandle(req.ContactInfo)
	if err != nil {
		return fmt.Errorf("normalized tg nickname:%w", err)
	}

	req.ContactInfo = normalized_nickname

	return s.repo.CreateApplication(ctx, req)
}

func (s *ApplicationsService) GetApplicationByID(ctx context.Context, id int64) (*models.Application, error) {
	return s.repo.GetApplicationByID(ctx, id)
}

func (s *ApplicationsService) UpdateApplicationStatus(ctx context.Context, id int64, status string) (*models.Application, error) {
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

	if err = s.repo.UpdateApplicationStatus(txCtx, id, status); err != nil {
		return nil, err
	}

	app, err := s.repo.GetApplicationByID(txCtx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *ApplicationsService) ApplicationsList(ctx context.Context, app *models.ApplicationFilter) (*models.ApplicationList, error) {
	return s.repo.ApplicationsList(ctx, app)
}

func (s *ApplicationsService) DeleteApplication(ctx context.Context, id int64) error {
	return s.repo.DeleteApplication(ctx, id)
}

func normalizeTelegramHandle(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", errors.New("telegram nickname is empty")
	}

	if idx := strings.Index(s, "t.me/"); idx != -1 {
		s = s[idx+len("t.me/"):]
	}

	if idx := strings.Index(s, "@"); idx != -1 {
		s = s[idx:]
	} else {
		s = "@" + s
	}

	handle := s[1:]
	if !isValidTelegramHandle(handle) {
		return "", fmt.Errorf("invalid telegram nickname %q: must be 4–32 chars, a-z, 0-9, underscore", handle)
	}

	return s, nil
}

func isValidTelegramHandle(handle string) bool {
	if len(handle) < 4 || len(handle) > 32 {
		return false
	}
	for _, r := range handle {
		if !isAllowedChar(r) {
			return false
		}
	}
	return true
}

func isAllowedChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}
