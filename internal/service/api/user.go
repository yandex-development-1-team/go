package service

import (
	"context"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

type UserService struct {
	repo repository.StaffRepository
}

func NewUserService(repo repository.StaffRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List(ctx context.Context, role, status, search string, limit, offset int) ([]dto.UserListItem, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, role, status, search, limit, offset)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*dto.UserWithDetails, error) {
	if id <= 0 {
		return nil, models.ErrInvalidInput
	}

	return s.repo.GetByID(ctx, id)
}

func (s *UserService) Dashboard(ctx context.Context, managerId int64) (*dto.DashboardResponse, error) {
	return s.repo.GetDashboard(ctx, managerId)
}
