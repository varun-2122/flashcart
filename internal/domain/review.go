package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReviewNotFound = errors.New("review not found")
	ErrInvalidRating  = errors.New("rating must be between 1 and 5")
)

type Review struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Optional joined fields for UI
	UserFirstName string `json:"user_first_name,omitempty"`
}

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*Review, error)
}
