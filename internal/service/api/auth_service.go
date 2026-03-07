package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yandex-development-1-team/go/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID               int64  `json:"uid"`
	Role                 string `json:"role"`
	jwt.RegisteredClaims        // exp, iat, iss, jti — стандартные поля
}

type UserApiRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.UserWithAuth, error)
	// GetAuthInfo(ctx context.Context, email string) (string, error)
}

type RefreshTokenRepository interface {
	CreateRefreshToken(context.Context, int64, string, time.Time) error
}

type AuthService struct {
	userRepo    UserApiRepository
	refreshRepo RefreshTokenRepository
}

func GetNewAuthService(userRepo UserApiRepository, refreshRepo RefreshTokenRepository) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
	}
}

func (as *AuthService) Login(ctx context.Context, email, password string) (*models.AuthResult, error) {
	// проверка на ошибочный ввод и атаки
	if len(email) == 0 || len(password) < 8 {
		return nil, models.ErrInvalidInput
	}

	authInfo, err := as.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	user := authInfo.User
	if err = CheckPassword(password, authInfo.PassHash); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, models.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("check password: %w", err)
	}

	if user.Status == "blocked" {
		return nil, models.ErrUserBlocked
	}

	token, err := generateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	err = as.refreshRepo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	return &models.AuthResult{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func HashPassword(password string) (string, error) {
	// DefultCost для баланса защиты и скрости работы
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func generateToken(userId int64, role string) (string, error) {
	claims := Claims{
		UserID: userId,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := jwtToken.SignedString([]byte("secret_key")) // заменить Secret на строку из конфига
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func generateRefreshToken() (string, error) {
	refreshTokenRaw := make([]byte, 32)

	_, err := rand.Read(refreshTokenRaw)
	if err != nil {
		return "", err
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(refreshTokenRaw)
	return refreshToken, nil
}
