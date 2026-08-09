package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

type RegisterRequest struct {
	Email     string      `json:"email"`
	Password  string      `json:"password"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Role      domain.Role `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User        *domain.User `json:"user"`
	AccessToken string       `json:"access_token"`
}

type AuthService struct {
	userRepo   domain.UserRepository
	jwtManager *JWTManager
}

func NewAuthService(userRepo domain.UserRepository, jwtManager *JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := req.Role
	if role == "" {
		role = domain.RoleCustomer
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:        user,
		AccessToken: token,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:        user,
		AccessToken: token,
	}, nil
}
