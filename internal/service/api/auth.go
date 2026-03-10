package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/yandex-development-1-team/go/internal/models"
	"github.com/yandex-development-1-team/go/internal/repository"
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
	userRepo   repository.UserRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type AccessClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sqlx.DB, rtRepo repository.RefreshTokenRepository, userRepo repository.UserRepository, jwtSecret string, accessTTLMinutes, refreshTTlDays int) *AuthService {
	return &AuthService{
		db:         db,
		rtRepo:     rtRepo,
		userRepo:   userRepo,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshTTlDays) * time.Hour * 24,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResult, error) {
	if len(email) == 0 || len(password) < 8 {
		return nil, models.ErrInvalidInput
	}
	authInfo, err := s.userRepo.GetUserByEmail(ctx, email)
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
	defer tx.Rollback()

	rt, err := s.rtRepo.GetForUpdate(ctx, tx, refreshToken)
	if err != nil {
		return "", err
	}

	role := "manager"

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
	return token.SignedString(s.jwtSecret)
}

func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
