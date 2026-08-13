package review

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/database"
	"github.com/varun-2122/flashcart/internal/domain"
)

type PostgresReviewRepository struct {
	db *database.PostgresDB
}

func NewPostgresReviewRepository(db *database.PostgresDB) domain.ReviewRepository {
	return &PostgresReviewRepository{db: db}
}

func (r *PostgresReviewRepository) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, product_id, user_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (product_id, user_id) 
		DO UPDATE SET rating = EXCLUDED.rating, comment = EXCLUDED.comment, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Pool.Exec(
		ctx, query,
		review.ID, review.ProductID, review.UserID, review.Rating, review.Comment, review.CreatedAt, review.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert or update review: %w", err)
	}
	return nil
}

func (r *PostgresReviewRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*domain.Review, error) {
	query := `
		SELECT r.id, r.product_id, r.user_id, r.rating, r.comment, r.created_at, r.updated_at, u.first_name
		FROM reviews r
		JOIN users u ON r.user_id = u.id
		WHERE r.product_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		rev := &domain.Review{}
		if err := rows.Scan(
			&rev.ID, &rev.ProductID, &rev.UserID, &rev.Rating, &rev.Comment, &rev.CreatedAt, &rev.UpdatedAt, &rev.UserFirstName,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}
