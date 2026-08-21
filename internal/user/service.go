package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

// UpdateProfileRequest holds editable user fields.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserService handles user profile operations.
type UserService struct {
	repo domain.UserRepository
}

// NewUserService creates a UserService.
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetProfile fetches a user by ID.
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// UpdateProfile patches editable user fields.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if req.FirstName != "" {
		u.FirstName = req.FirstName
	}
	if req.LastName != "" {
		u.LastName = req.LastName
	}
	u.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return u, nil
}
