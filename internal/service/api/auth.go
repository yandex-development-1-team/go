package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const (
	RoleAdmin    = "admin"
	RoleManager1 = "manager_1"
	RoleManager2 = "manager_2"
	RoleManager3 = "manager_3"
	RoleUser     = "user"
)

var Roles = []string{RoleAdmin, RoleManager1, RoleManager2, RoleManager3}

// HashPassword hashes a password for storage (e.g. when creating/updating users).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

type AuthService struct {
	db         *sqlx.DB
	rtRepo     repository.RefreshTokenRepository
	prRepo     repository.PasswordResetRepository
	staffRepo  repository.StaffRepository
	emailSvc   *EmailService
	JwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	txRepo     repository.TxRepository
}

type AccessClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type ResetPasswordClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sqlx.DB, rtRepo repository.RefreshTokenRepository, prRepo repository.PasswordResetRepository, staffRepo repository.StaffRepository,
	emailSvc *EmailService, txRepo repository.TxRepository, jwtSecret string, accessTTLMinutes, refreshTTlDays int) *AuthService {
	return &AuthService{
		db:         db,
		rtRepo:     rtRepo,
		prRepo:     prRepo,
		staffRepo:  staffRepo,
		emailSvc:   emailSvc,
		JwtSecret:  []byte(jwtSecret),
		txRepo:     txRepo,
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTlDays) * time.Hour * 24,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResult, error) {
	authInfo, err := s.staffRepo.GetUserByEmail(ctx, email)
	if errors.Is(err, models.ErrUserNotFound) {
		return nil, models.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	user := authInfo.User
	if user.Status == "blocked" {
		return nil, models.ErrUserBlocked
	}
	if err := bcrypt.CompareHashAndPassword([]byte(authInfo.PassHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrHashTooShort) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("compare password hash: %w", err)
	}
	token, err := s.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	err = s.rtRepo.DeleteByStaffID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken := generateRandomToken()
	expiresAt := time.Now().Add(s.refreshTTL)
	if err := s.rtRepo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &models.AuthResult{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*models.RefreshResponse, error) {
	var accessToken string
	var newRT *models.RefreshToken

	err := s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {
		rt, err := s.rtRepo.GetForUpdate(txCtx, refreshToken)
		if err != nil {
			return err
		}

		accessToken, err = s.generateAccessToken(rt.UserID, rt.Role)
		if err != nil {
			return err
		}

		newRT = &models.RefreshToken{
			UserID:    rt.UserID,
			Token:     generateRandomToken(),
			ExpiresAt: time.Now().Add(s.refreshTTL),
		}
		if err := s.rtRepo.Create(txCtx, newRT); err != nil {
			return err
		}

		if err := s.rtRepo.DeleteByToken(txCtx, refreshToken); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &models.RefreshResponse{
		Token:        accessToken,
		RefreshToken: newRT.Token,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, user *models.UserAPI, password string) (*models.AuthResult, error) {
	hashPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	refreshToken := generateRandomToken()
	expiresAt := time.Now().Add(s.refreshTTL)

	var accessToken string
	err = s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {

		user, err = s.staffRepo.CreateStaff(txCtx, user, hashPassword)
		if err != nil {
			return err
		}

		err = s.rtRepo.CreateRefreshToken(txCtx, user.ID, refreshToken, expiresAt)
		if err != nil {
			return err
		}

		// accessToken
		accessToken, err = s.generateAccessToken(user.ID, user.Role)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// в случае успеха коммит и возвращаем User
	return &models.AuthResult{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.rtRepo.DeleteByToken(ctx, refreshToken)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || !domainHasMX(parts[1]) {
		return models.ErrInvalidEmail
	}

	var signedToken string
	err := s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {
		user, err := s.staffRepo.GetUserByEmail(txCtx, email)
		if err != nil {
			return err
		}

		// Генерируем JWT токен для сброса пароля
		resetClaims := ResetPasswordClaims{
			UserID: user.User.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)
		signedToken, err = token.SignedString(s.JwtSecret)
		if err != nil {
			return fmt.Errorf("sign reset token: %w", err)
		}

		expiresAt := time.Now().Add(1 * time.Hour)

		err = s.prRepo.CreateToken(txCtx, user.User.ID, signedToken, expiresAt)
		if err != nil {
			return fmt.Errorf("create reset token: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Send email
	if err = s.emailSvc.SendPasswordResetEmail(ctx, email, signedToken); err != nil {
		var sendErr *mail.SendError
		if errors.As(err, &sendErr) {
			var textErr *textproto.Error
			if errors.As(sendErr.Cause, &textErr) && textErr.Code >= 550 && textErr.Code <= 553 {
				return nil
			}
		}
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	var claims ResetPasswordClaims
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return s.JwtSecret, nil
	})
	if err != nil || !parsedToken.Valid {
		return models.ErrInvalidCredentials
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.txRepo.RunToTx(ctx, func(txCtx context.Context) error {
		prt, err := s.prRepo.GetToken(txCtx, token)
		if err != nil {
			if errors.Is(err, models.ErrTokenNotFound) {
				return models.ErrInvalidCredentials
			}
			return err
		}

		if prt.UsedAt != nil {
			return models.ErrInvalidCredentials
		}

		if err := s.staffRepo.UpdatePassword(txCtx, prt.UserID, hash); err != nil {
			return err
		}

		if err := s.prRepo.DeleteToken(txCtx, prt.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *AuthService) generateAccessToken(userID int64, role string) (string, error) {
	now := time.Now().UTC()

	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JwtSecret)
}

func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func domainHasMX(domain string) bool {
	records, err := net.LookupMX(domain)
	return err == nil && len(records) > 0
}
