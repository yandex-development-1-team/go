package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
)

const (
	RoleAdmin    = "admin"
	RoleManager1 = "manager1"
	RoleManager2 = "manager2"
	RoleManager3 = "manager3"
)

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
	if len(email) == 0 || len(password) < 8 {
		return nil, models.ErrInvalidInput
	}
	authInfo, err := s.staffRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	user := authInfo.User
	if err := bcrypt.CompareHashAndPassword([]byte(authInfo.PassHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("check password: %w", err)
	}
	if user.Status == "blocked" {
		return nil, models.ErrUserBlocked
	}
	token, err := s.generateAccessToken(user.ID, user.Role)
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

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	rt, err := s.rtRepo.GetForUpdate(ctx, tx, refreshToken)
	if err != nil {
		return "", err
	}

	role := RoleManager1

	accessToken, err := s.generateAccessToken(rt.UserID, role)
	if err != nil {
		return "", err
	}

	newRT := &models.RefreshToken{
		UserID:    rt.UserID,
		Token:     generateRandomToken(),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.rtRepo.Create(ctx, newRT); err != nil {
		return "", err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token = $1 AND revoked_at IS NULL
	`, refreshToken); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return accessToken, nil
}

func (s *AuthService) Register(ctx context.Context, user *models.UserAPI, password string) (*models.AuthResult, error) {
	hashPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	refreshToken := generateRandomToken()
	expiresAt := time.Now().Add(s.refreshTTL)

	err = s.txRepo.RunToTx(ctx, func(ctx context.Context) error {

		user, err = s.staffRepo.CreateStaff(ctx, user, hashPassword)
		if err != nil {
			return err
		}

		err = s.rtRepo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// accessToken
	accessToken, err := s.generateAccessToken(user.ID, user.Role)
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
	return s.rtRepo.Revoke(ctx, refreshToken)
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

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.staffRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Не раскрываем, существует ли email
		return nil
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
	signedToken, err := token.SignedString(s.JwtSecret)
	if err != nil {
		return fmt.Errorf("sign reset token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)

	err = s.prRepo.CreateToken(ctx, user.User.ID, signedToken, expiresAt)
	if err != nil {
		return fmt.Errorf("create reset token: %w", err)
	}

	// Send email
	err = s.emailSvc.SendPasswordResetEmail(ctx, email, signedToken)
	if err != nil {
		log.Printf("Failed to send password reset email: %v", err)
		// Не возвращаем ошибку
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Парсим JWT токен
	var claims ResetPasswordClaims
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return s.JwtSecret, nil
	})
	if err != nil || !parsedToken.Valid {
		return models.ErrInvalidCredentials
	}

	// Проверяем, что токен в БД
	prt, err := s.prRepo.GetToken(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrTokenNotFound) {
			return models.ErrInvalidCredentials
		}
		return err
	}

	if prt.UsedAt != nil {
		return models.ErrInvalidCredentials
	}

	if time.Now().After(prt.ExpiresAt) {
		return models.ErrInvalidCredentials
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Update password and remove reset token in a transaction
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	if _, err := tx.ExecContext(ctx, query, hash, time.Now(), prt.UserID); err != nil {
		return err
	}

	query = `DELETE FROM password_reset_tokens WHERE id = $1`
	if _, err := tx.ExecContext(ctx, query, prt.ID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
