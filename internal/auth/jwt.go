package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Claims represents JWT token claims.
type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager signs and validates JWT tokens.
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager initializes JWTManager.
func NewJWTManager(secret string, duration time.Duration) *JWTManager {
	if secret == "" {
		secret = "flashcart_production_secret_key_change_me_in_env"
	}
	return &JWTManager{
		secretKey:     []byte(secret),
		tokenDuration: duration,
	}
}

// GenerateToken creates a signed JWT token.
func (m *JWTManager) GenerateToken(user *domain.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// VerifyToken parses and validates a JWT token string.
func (m *JWTManager) VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.secretKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
