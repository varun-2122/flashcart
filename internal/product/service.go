package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
)

type CreateProductRequest struct {
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Brand       string     `json:"brand"`
	Quantity    int        `json:"quantity"` // Initial stock
}

type ProductService struct {
	repo          domain.ProductRepository
	inventoryRepo domain.InventoryRepository
}

func NewProductService(repo domain.ProductRepository, inventoryRepo domain.InventoryRepository) *ProductService {
	return &ProductService{
		repo:          repo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req CreateProductRequest) (*domain.Product, error) {
	if req.SKU == "" || req.Name == "" || req.Price < 0 {
		return nil, domain.ErrInvalidSKU
	}

	now := time.Now()
	product := &domain.Product{
		ID:          uuid.New(),
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Initialize inventory stock if repository provided
	if s.inventoryRepo != nil {
		_ = s.inventoryRepo.SetStock(ctx, product.ID, req.Quantity)
	}

	return product, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error) {
	return s.repo.List(ctx, filter)
}
