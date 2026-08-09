package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error hashing password, got %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Errorf("expected CheckPassword to return true for correct password")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Errorf("expected CheckPassword to return false for incorrect password")
	}
}

func TestJWTManager(t *testing.T) {
	jwtManager := NewJWTManager("test_secret_key", 1*time.Hour)

	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@flashcart.com",
		Role:  domain.RoleCustomer,
	}

	token, err := jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	claims, err := jwtManager.VerifyToken(token)
	if err != nil {
		t.Fatalf("expected no error verifying token, got %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != domain.RoleCustomer {
		t.Errorf("expected role customer, got %s", claims.Role)
	}
}
