package review

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

type ReviewService struct {
	repo domain.ReviewRepository
}

func NewReviewService(repo domain.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) CreateReview(ctx context.Context, productID, userID uuid.UUID, rating int, comment string) (*domain.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, domain.ErrInvalidRating
	}

	review := &domain.Review{
		ID:        uuid.New(),
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (s *ReviewService) GetProductReviews(ctx context.Context, productID uuid.UUID) ([]*domain.Review, error) {
	return s.repo.GetByProductID(ctx, productID)
}
