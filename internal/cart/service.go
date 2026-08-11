package cart

import (
	"context"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/metrics"
)

type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type CartService struct {
	cartRepo    domain.CartRepository
	productRepo domain.ProductRepository
}

func NewCartService(cartRepo domain.CartRepository, productRepo domain.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, req AddItemRequest) (*domain.Cart, error) {
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		cart = &domain.Cart{UserID: userID, Items: []domain.CartItem{}}
	}

	found := false
	for i, item := range cart.Items {
		if item.ProductID == req.ProductID {
			cart.Items[i].Quantity += req.Quantity
			found = true
			break
		}
	}

	if !found {
		cart.Items = append(cart.Items, domain.CartItem{
			ProductID: product.ID,
			Name:      product.Name,
			UnitPrice: product.Price,
			Quantity:  req.Quantity,
		})
	}

	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	// Track cart engagement metric for new unique items only
	if !found {
		metrics.CartItemsAdded.Inc()
	}

	return cart, nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return cart, nil
	}

	newItems := make([]domain.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems
	if err := s.cartRepo.Save(ctx, cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*domain.Cart, error) {
	return s.cartRepo.GetByUserID(ctx, userID)
}
