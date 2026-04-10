package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Thorlik/marketplace/internal/model"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrCartEmpty          = errors.New("cart is empty")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrProductUnavailable = errors.New("product is unavailable")
	ErrNotOrderOwner      = errors.New("you are not the owner of this order")
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error)
	CreateOrderFromCart(ctx context.Context, userID int64) (*domain.Order, error)
	GetOrderByID(ctx context.Context, userID int64, userRole domain.UserRole, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Order, error)
	GetAllOrders(ctx context.Context, page, pageSize int) ([]*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error
	ProcessPayment(ctx context.Context, userID int64, orderID int64, paymentMethod string) (*PaymentResult, error)

	AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error)
	GetCart(ctx context.Context, userID int64) ([]*domain.CartItem, float64, error)
	UpdateCartItem(ctx context.Context, userID, cartItemID int64, quantity int) error
	RemoveFromCart(ctx context.Context, userID, cartItemID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type PaymentResult struct {
	PaymentID     string
	Status        string
	Message       string
	TransactionID string
}

type orderService struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	logger      *zap.Logger
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, logger *zap.Logger) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	s.logger.Info("Creating order", zap.Int64("user_id", userID), zap.Int("items_count", len(items)))

	if len(items) == 0 {
		return nil, ErrCartEmpty
	}

	var totalAmount float64
	for i := range items {
		product, err := s.productRepo.GetByID(ctx, items[i].ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				return nil, fmt.Errorf("product %d not found", items[i].ProductID)
			}
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		if product.Status != "approved" {
			return nil, ErrProductUnavailable
		}

		if product.Quantity < items[i].Quantity {
			return nil, ErrInsufficientStock
		}

		items[i].PriceAtPurchase = product.Price
		totalAmount += product.Price * float64(items[i].Quantity)
	}

	order := &domain.Order{
		UserID:        userID,
		Status:        domain.OrderStatusPending,
		TotalAmount:   totalAmount,
		PaymentStatus: domain.PaymentStatusPending,
	}

	entityOrder := model.OrderDomainToEntity(order)
	createdOrder, err := s.orderRepo.CreateOrder(ctx, entityOrder)
	if err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	for _, item := range items {
		entityItem := model.OrderItemDomainToEntity(&domain.OrderItem{
			OrderID:         createdOrder.ID,
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			PriceAtPurchase: item.PriceAtPurchase,
		})

		if _, err := s.orderRepo.CreateOrderItem(ctx, entityItem); err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		product, _ := s.productRepo.GetByID(ctx, item.ProductID)
		newQuantity := product.Quantity - item.Quantity
		if err := s.productRepo.UpdateQuantity(ctx, item.ProductID, newQuantity); err != nil {
			s.logger.Error("Failed to update product quantity", zap.Error(err))
		}
	}

	s.logger.Info("Order created successfully", zap.Int64("id", createdOrder.ID))

	resultOrder := model.OrderEntityToDomain(createdOrder)
	resultOrder.Items = items

	return resultOrder, nil
}

func (s *orderService) CreateOrderFromCart(ctx context.Context, userID int64) (*domain.Order, error) {
	s.logger.Info("Creating order from cart", zap.Int64("user_id", userID))

	cartItems, err := s.orderRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, ErrCartEmpty
	}

	items := make([]domain.OrderItem, 0, len(cartItems))
	for _, ci := range cartItems {
		items = append(items, domain.OrderItem{
			ProductID: ci.ProductID,
			Quantity:  ci.Quantity,
		})
	}

	order, err := s.CreateOrder(ctx, userID, items)
	if err != nil {
		return nil, err
	}

	if err := s.orderRepo.ClearCart(ctx, userID); err != nil {
		s.logger.Error("Failed to clear cart", zap.Error(err))
	}

	return order, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, userID int64, userRole domain.UserRole, orderID int64) (*domain.Order, error) {
	entityOrder, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if entityOrder.UserID != userID && userRole != domain.RoleAdmin {
		return nil, ErrNotOrderOwner
	}

	order := model.OrderEntityToDomain(entityOrder)

	entityItems, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		s.logger.Error("Failed to get order items", zap.Error(err))
	} else {
		for _, ei := range entityItems {
			order.Items = append(order.Items, *model.OrderItemEntityToDomain(ei))
		}
	}

	return order, nil
}

func (s *orderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Order, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	entityOrders, err := s.orderRepo.GetOrdersByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]*domain.Order, 0, len(entityOrders))
	for _, eo := range entityOrders {
		orders = append(orders, model.OrderEntityToDomain(eo))
	}

	return orders, nil
}

func (s *orderService) GetAllOrders(ctx context.Context, page, pageSize int) ([]*domain.Order, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	entityOrders, err := s.orderRepo.GetAllOrders(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]*domain.Order, 0, len(entityOrders))
	for _, eo := range entityOrders {
		orders = append(orders, model.OrderEntityToDomain(eo))
	}

	return orders, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	s.logger.Info("Updating order status", zap.Int64("id", orderID), zap.String("status", string(status)))

	if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, string(status)); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *orderService) ProcessPayment(ctx context.Context, userID int64, orderID int64, paymentMethod string) (*PaymentResult, error) {
	s.logger.Info("Processing payment", zap.Int64("order_id", orderID), zap.String("method", paymentMethod))

	order, err := s.orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return nil, ErrNotOrderOwner
	}

	// пока оставил просто временную заглушку как было посоветовано в МР
	paymentID := fmt.Sprintf("PAY-%d-%d", orderID, userID)
	transactionID := fmt.Sprintf("TXN-%d", orderID)

	result := &PaymentResult{
		PaymentID:     paymentID,
		Status:        "success",
		Message:       "Payment processed successfully",
		TransactionID: transactionID,
	}

	if err := s.orderRepo.UpdatePaymentStatus(ctx, orderID, string(domain.PaymentStatusSuccess), paymentID); err != nil {
		s.logger.Error("Failed to update payment status", zap.Error(err))
	}

	if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, string(domain.OrderStatusPaid)); err != nil {
		s.logger.Error("Failed to update order status", zap.Error(err))
	}

	s.logger.Info("Payment processed successfully", zap.String("payment_id", paymentID))
	return result, nil
}

func (s *orderService) AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error) {
	s.logger.Debug("Adding to cart", zap.Int64("user_id", userID), zap.Int64("product_id", productID))

	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if product.Status != "approved" {
		return nil, ErrProductUnavailable
	}

	if product.Quantity < quantity {
		return nil, ErrInsufficientStock
	}

	entityItem := &entity.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}

	createdItem, err := s.orderRepo.AddToCart(ctx, entityItem)
	if err != nil {
		return nil, fmt.Errorf("failed to add to cart: %w", err)
	}

	return model.CartItemEntityToDomain(createdItem), nil
}

func (s *orderService) GetCart(ctx context.Context, userID int64) ([]*domain.CartItem, float64, error) {
	entityItems, err := s.orderRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cart: %w", err)
	}

	var totalPrice float64
	items := make([]*domain.CartItem, 0, len(entityItems))

	for _, ei := range entityItems {
		item := model.CartItemEntityToDomain(ei)

		product, err := s.productRepo.GetByID(ctx, ei.ProductID)
		if err == nil {
			item.Product = model.ProductEntityToDomain(product)
			totalPrice += product.Price * float64(ei.Quantity)
		}

		items = append(items, item)
	}

	return items, totalPrice, nil
}

func (s *orderService) UpdateCartItem(ctx context.Context, userID, cartItemID int64, quantity int) error {
	return s.orderRepo.UpdateCartItem(ctx, cartItemID, quantity)
}

func (s *orderService) RemoveFromCart(ctx context.Context, userID, cartItemID int64) error {
	return s.orderRepo.DeleteCartItem(ctx, cartItemID)
}

func (s *orderService) ClearCart(ctx context.Context, userID int64) error {
	return s.orderRepo.ClearCart(ctx, userID)
}
