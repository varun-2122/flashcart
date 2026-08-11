package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/varun-2122/flashcart/internal/domain"
	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
	"github.com/varun-2122/flashcart/internal/worker"
)

type OrderService struct {
	orderRepo     domain.OrderRepository
	cartRepo      domain.CartRepository
	inventoryRepo domain.InventoryRepository
	productRepo   domain.ProductRepository
	workerPool    *worker.Pool
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	cartRepo domain.CartRepository,
	inventoryRepo domain.InventoryRepository,
	productRepo domain.ProductRepository,
	workerPool *worker.Pool,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
		workerPool:    workerPool,
	}
}

// CreateOrderFromCart processes checkout using Optimistic Locking and Unit of Work logic.
func (s *OrderService) CreateOrderFromCart(ctx context.Context, userID uuid.UUID) (*domain.Order, error) {
	cart, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil || cart == nil || len(cart.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	// 1. Reserve Inventory using Optimistic Locking for each item
	orderItems := make([]domain.OrderItem, 0, len(cart.Items))
	now := time.Now()
	orderID := uuid.New()
	var totalAmount float64

	for _, item := range cart.Items {
		if s.inventoryRepo != nil {
			inv, err := s.inventoryRepo.GetByProductID(ctx, item.ProductID)
			if err != nil {
				// If no inventory record exists yet, allow demo checkout or check stock
				logger.Warn(ctx, "inventory record not found during checkout, skipping optimistic lock probe", "product_id", item.ProductID.String())
			} else {
				if inv.AvailableQuantity() < item.Quantity {
					return nil, fmt.Errorf("insufficient stock for product %s: %w", item.Name, domain.ErrInsufficientStock)
				}
				// Atomic Optimistic Lock Update
				if err := s.inventoryRepo.ReserveStockWithOptimisticLock(ctx, item.ProductID, item.Quantity, inv.Version); err != nil {
					return nil, fmt.Errorf("failed to reserve inventory for %s: %w", item.Name, err)
				}
			}
		}

		itemTotal := item.UnitPrice * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, domain.OrderItem{
			ID:        uuid.New(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			CreatedAt: now,
		})
	}

	// 2. Build Order Entity
	order := &domain.Order{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      domain.OrderStatusPending,
		Items:       orderItems,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 3. Persist Order in Database
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to persist order: %w", err)
	}

	// 4. Clear Shopping Cart
	_ = s.cartRepo.Clear(ctx, userID)

	// 5. Record Business Metrics
	metrics.OrdersCreated.Inc()
	metrics.OrderTotalAmount.Observe(totalAmount)

	logger.Info(ctx, "order created successfully", "order_id", orderID.String(), "user_id", userID.String(), "total", totalAmount)

	// 6. Dispatch Async Post-Checkout Jobs (non-blocking)
	if s.workerPool != nil {
		s.workerPool.Dispatch(ctx, &worker.OrderCreatedJob{
			OrderID:   orderID,
			UserID:    userID,
			Total:     totalAmount,
			CreatedAt: now,
		})
		s.workerPool.Dispatch(ctx, &worker.InvoiceJob{
			OrderID: orderID,
			UserID:  userID,
			Total:   totalAmount,
		})
	}

	return order, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListUserOrders(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID)
}
