package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"google.golang.org/api/idtoken"
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
	userRepo       domain.UserRepository
	jwtManager     *JWTManager
	googleClientID string
}

func NewAuthService(userRepo domain.UserRepository, jwtManager *JWTManager, googleClientID string) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		googleClientID: googleClientID,
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

func (s *AuthService) GoogleLogin(ctx context.Context, idToken string) (*AuthResponse, error) {
	payload, err := idtoken.Validate(ctx, idToken, s.googleClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google id_token: %w", err)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email not found in google token")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Auto-register new user
			firstName, _ := payload.Claims["given_name"].(string)
			lastName, _ := payload.Claims["family_name"].(string)

			// Generate a random password hash since they won't use passwords
			randomPass := uuid.New().String()
			hashedPassword, _ := HashPassword(randomPass)
			
			now := time.Now()
			user = &domain.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: hashedPassword,
				FirstName:    firstName,
				LastName:     lastName,
				Role:         domain.RoleCustomer,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to auto-register google user: %w", err)
			}
		} else {
			return nil, err
		}
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
