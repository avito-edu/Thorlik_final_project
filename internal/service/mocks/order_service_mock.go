package mocks

import (
	"context"

	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

func NewMockOrderService() *MockOrderService {
	return &MockOrderService{}
}

func (m *MockOrderService) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	args := m.Called(ctx, userID, items)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) CreateOrderFromCart(ctx context.Context, userID int64) (*domain.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetOrderByID(ctx context.Context, userID int64, userRole domain.UserRole, orderID int64) (*domain.Order, error) {
	args := m.Called(ctx, userID, userRole, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetUserOrders(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Order, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderService) GetAllOrders(ctx context.Context, page, pageSize int) ([]*domain.Order, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrderService) ProcessPayment(ctx context.Context, userID int64, orderID int64, paymentMethod string) (*service.PaymentResult, error) {
	args := m.Called(ctx, userID, orderID, paymentMethod)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.PaymentResult), args.Error(1)
}

func (m *MockOrderService) AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error) {
	args := m.Called(ctx, userID, productID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CartItem), args.Error(1)
}

func (m *MockOrderService) GetCart(ctx context.Context, userID int64) ([]*domain.CartItem, float64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(float64), args.Error(2)
	}
	return args.Get(0).([]*domain.CartItem), args.Get(1).(float64), args.Error(2)
}

func (m *MockOrderService) UpdateCartItem(ctx context.Context, userID, cartItemID int64, quantity int) error {
	args := m.Called(ctx, userID, cartItemID, quantity)
	return args.Error(0)
}

func (m *MockOrderService) RemoveFromCart(ctx context.Context, userID, cartItemID int64) error {
	args := m.Called(ctx, userID, cartItemID)
	return args.Error(0)
}

func (m *MockOrderService) ClearCart(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
