package mocks

import (
	"context"

	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/stretchr/testify/mock"
)

type MockOrderRepository struct {
	mock.Mock
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{}
}

func (m *MockOrderRepository) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Order, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]*entity.Order, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdatePaymentStatus(ctx context.Context, id int64, status, paymentID string) error {
	args := m.Called(ctx, id, status, paymentID)
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrderItem(ctx context.Context, item *entity.OrderItem) (*entity.OrderItem, error) {
	args := m.Called(ctx, item)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OrderItem), args.Error(1)
}

func (m *MockOrderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]*entity.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrderItem), args.Error(1)
}

func (m *MockOrderRepository) AddToCart(ctx context.Context, item *entity.CartItem) (*entity.CartItem, error) {
	args := m.Called(ctx, item)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockOrderRepository) GetCartByUserID(ctx context.Context, userID int64) ([]*entity.CartItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.CartItem), args.Error(1)
}

func (m *MockOrderRepository) GetCartItem(ctx context.Context, userID, productID int64) (*entity.CartItem, error) {
	args := m.Called(ctx, userID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CartItem), args.Error(1)
}

func (m *MockOrderRepository) UpdateCartItem(ctx context.Context, id int64, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockOrderRepository) DeleteCartItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) ClearCart(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
