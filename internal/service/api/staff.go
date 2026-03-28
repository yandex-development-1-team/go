package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/yandex-development-1-team/go/internal/dto"
	"github.com/yandex-development-1-team/go/internal/models"
	repo "github.com/yandex-development-1-team/go/internal/repository/postgres"
)

type UsersAdminService struct {
	staffRepo *repo.StaffRepo
	rtRepo    *repo.RefreshTokenRepo
}

func NewUsersAdminService(staffRepo *repo.StaffRepo, rtRepo *repo.RefreshTokenRepo) *UsersAdminService {
	return &UsersAdminService{staffRepo: staffRepo, rtRepo: rtRepo}
}

func (s *UsersAdminService) Create(ctx context.Context, req dto.UserCreateRequest) (*models.UserAPI, error) {
	status := req.Status
	if status == "" {
		status = "invited"
	}
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	if err := validateStaffEnums(req.Role, status); err != nil {
		return nil, err
	}

	m := &models.StaffAdminCreate{
		Name:         req.Name,
		Email:        req.Email,
		Role:         req.Role,
		Status:       status,
		Permissions:  perms,
		TelegramNick: req.TelegramNick,
		InviteToken:  generateInviteToken(),
	}
	return s.staffRepo.CreateStaffByAdmin(ctx, m)
}

func (s *UsersAdminService) Update(ctx context.Context, id int64, req dto.UserUpdateRequest) (*models.UserAPI, error) {
	u := &models.StaffAdminUpdate{
		Name:         req.Name,
		Email:        req.Email,
		Role:         req.Role,
		Status:       req.Status,
		Permissions:  req.Permissions,
		TelegramNick: req.TelegramNick,
	}
	if u.Role != nil {
		if err := validateRole(*u.Role); err != nil {
			return nil, err
		}
	}
	if u.Status != nil {
		if err := validateStatus(*u.Status); err != nil {
			return nil, err
		}
	}
	return s.staffRepo.UpdateStaff(ctx, id, u)
}

func (s *UsersAdminService) Block(ctx context.Context, id int64) (*models.UserAPI, error) {
	user, err := s.staffRepo.BlockStaff(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.rtRepo.RevokeByUserID(ctx, id); err != nil {
		return nil, err
	}
	return user, nil
}

func validateStaffEnums(role, status string) error {
	if err := validateRole(role); err != nil {
		return err
	}
	return validateStatus(status)
}

func validateRole(role string) error {
	switch role {
	case "admin", "manager_1", "manager_2", "manager_3", "user":
		return nil
	default:
		return models.ErrInvalidInput
	}
}

func validateStatus(status string) error {
	switch status {
	case "active", "blocked", "invited":
		return nil
	default:
		return models.ErrInvalidInput
	}
}

func generateInviteToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
