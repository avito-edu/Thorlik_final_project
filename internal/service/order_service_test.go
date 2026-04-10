package service

import (
	"context"
	"testing"
	"time"

	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/repository"
	"github.com/Thorlik/marketplace/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestOrderService(orderRepo *mocks.MockOrderRepository, productRepo *mocks.MockProductRepository) OrderService {
	logger := zap.NewNop()
	return NewOrderService(orderRepo, productRepo, logger)
}

func TestOrderService_CreateOrder_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
	}

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "approved",
	}, nil)
	productRepo.On("GetByID", ctx, int64(2)).Return(&entity.Product{
		ID:       2,
		Price:    50.0,
		Quantity: 5,
		Status:   "approved",
	}, nil)

	orderRepo.On("CreateOrder", ctx, mock.AnythingOfType("*entity.Order")).Return(&entity.Order{
		ID:            1,
		UserID:        1,
		Status:        "pending",
		TotalAmount:   250.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	orderRepo.On("CreateOrderItem", ctx, mock.AnythingOfType("*entity.OrderItem")).Return(&entity.OrderItem{
		ID:      1,
		OrderID: 1,
	}, nil).Twice()

	productRepo.On("UpdateQuantity", ctx, int64(1), 8).Return(nil)
	productRepo.On("UpdateQuantity", ctx, int64(2), 4).Return(nil)

	order, err := svc.CreateOrder(ctx, 1, items)

	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(1), order.ID)
	assert.Equal(t, 250.0, order.TotalAmount)
	orderRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_EmptyCart(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	order, err := svc.CreateOrder(ctx, 1, []domain.OrderItem{})

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrCartEmpty)
}

func TestOrderService_CreateOrder_ProductNotFound(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 999, Quantity: 1},
	}

	productRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrProductNotFound)

	order, err := svc.CreateOrder(ctx, 1, items)

	assert.Nil(t, order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product 999 not found")
	productRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_ProductUnavailable(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 1},
	}

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "pending",
	}, nil)

	order, err := svc.CreateOrder(ctx, 1, items)

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrProductUnavailable)
	productRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrder_InsufficientStock(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 100},
	}

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "approved",
	}, nil)

	order, err := svc.CreateOrder(ctx, 1, items)

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrInsufficientStock)
	productRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrderFromCart_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetCartByUserID", ctx, int64(1)).Return([]*entity.CartItem{
		{ID: 1, UserID: 1, ProductID: 1, Quantity: 2},
	}, nil)

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "approved",
	}, nil)

	orderRepo.On("CreateOrder", ctx, mock.AnythingOfType("*entity.Order")).Return(&entity.Order{
		ID:            1,
		UserID:        1,
		Status:        "pending",
		TotalAmount:   200.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	orderRepo.On("CreateOrderItem", ctx, mock.AnythingOfType("*entity.OrderItem")).Return(&entity.OrderItem{
		ID:      1,
		OrderID: 1,
	}, nil)

	productRepo.On("UpdateQuantity", ctx, int64(1), 8).Return(nil)
	orderRepo.On("ClearCart", ctx, int64(1)).Return(nil)

	order, err := svc.CreateOrderFromCart(ctx, 1)

	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, 200.0, order.TotalAmount)
	orderRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
}

func TestOrderService_CreateOrderFromCart_EmptyCart(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetCartByUserID", ctx, int64(1)).Return([]*entity.CartItem{}, nil)

	order, err := svc.CreateOrderFromCart(ctx, 1)

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrCartEmpty)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_Success_Owner(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(1)).Return(&entity.Order{
		ID:            1,
		UserID:        1,
		Status:        "pending",
		TotalAmount:   100.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	orderRepo.On("GetOrderItems", ctx, int64(1)).Return([]*entity.OrderItem{
		{ID: 1, OrderID: 1, ProductID: 1, Quantity: 2, PriceAtPurchase: 50.0},
	}, nil)

	order, err := svc.GetOrderByID(ctx, 1, domain.RoleUser, 1)

	require.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(1), order.ID)
	assert.Len(t, order.Items, 1)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_Success_Admin(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(1)).Return(&entity.Order{
		ID:            1,
		UserID:        2,
		Status:        "pending",
		TotalAmount:   100.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	orderRepo.On("GetOrderItems", ctx, int64(1)).Return([]*entity.OrderItem{}, nil)

	order, err := svc.GetOrderByID(ctx, 999, domain.RoleAdmin, 1)

	require.NoError(t, err)
	assert.NotNil(t, order)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_NotOwner(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(1)).Return(&entity.Order{
		ID:            1,
		UserID:        2,
		Status:        "pending",
		TotalAmount:   100.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	order, err := svc.GetOrderByID(ctx, 1, domain.RoleUser, 1)

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrNotOrderOwner)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetOrderByID_NotFound(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(999)).Return(nil, repository.ErrOrderNotFound)

	order, err := svc.GetOrderByID(ctx, 1, domain.RoleUser, 999)

	assert.Nil(t, order)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetUserOrders_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrdersByUserID", ctx, int64(1), 20, 0).Return([]*entity.Order{
		{ID: 1, UserID: 1, Status: "pending", TotalAmount: 100, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, UserID: 1, Status: "paid", TotalAmount: 200, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil)

	orders, err := svc.GetUserOrders(ctx, 1, 1, 20)

	require.NoError(t, err)
	assert.Len(t, orders, 2)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetUserOrders_InvalidPagination(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrdersByUserID", ctx, int64(1), 20, 0).Return([]*entity.Order{}, nil)

	_, err := svc.GetUserOrders(ctx, 1, -1, 200)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_GetAllOrders_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetAllOrders", ctx, 20, 0).Return([]*entity.Order{
		{ID: 1, UserID: 1, Status: "pending", TotalAmount: 100, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil)

	orders, err := svc.GetAllOrders(ctx, 1, 20)

	require.NoError(t, err)
	assert.Len(t, orders, 1)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateOrderStatus_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("UpdateOrderStatus", ctx, int64(1), "shipped").Return(nil)

	err := svc.UpdateOrderStatus(ctx, 1, domain.OrderStatusShipped)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateOrderStatus_NotFound(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("UpdateOrderStatus", ctx, int64(999), "shipped").Return(repository.ErrOrderNotFound)

	err := svc.UpdateOrderStatus(ctx, 999, domain.OrderStatusShipped)

	assert.ErrorIs(t, err, ErrOrderNotFound)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_ProcessPayment_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(1)).Return(&entity.Order{
		ID:            1,
		UserID:        1,
		Status:        "pending",
		TotalAmount:   100.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	orderRepo.On("UpdatePaymentStatus", ctx, int64(1), "success", mock.AnythingOfType("string")).Return(nil)
	orderRepo.On("UpdateOrderStatus", ctx, int64(1), "paid").Return(nil)

	result, err := svc.ProcessPayment(ctx, 1, 1, "card")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result.Status)
	assert.NotEmpty(t, result.PaymentID)
	assert.NotEmpty(t, result.TransactionID)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_ProcessPayment_NotOwner(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(1)).Return(&entity.Order{
		ID:            1,
		UserID:        2,
		Status:        "pending",
		TotalAmount:   100.0,
		PaymentStatus: "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil)

	result, err := svc.ProcessPayment(ctx, 1, 1, "card")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotOrderOwner)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_ProcessPayment_OrderNotFound(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetOrderByID", ctx, int64(999)).Return(nil, repository.ErrOrderNotFound)

	result, err := svc.ProcessPayment(ctx, 1, 999, "card")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrOrderNotFound)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_AddToCart_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "approved",
	}, nil)

	orderRepo.On("AddToCart", ctx, mock.AnythingOfType("*entity.CartItem")).Return(&entity.CartItem{
		ID:        1,
		UserID:    1,
		ProductID: 1,
		Quantity:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	item, err := svc.AddToCart(ctx, 1, 1, 2)

	require.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, int64(1), item.ProductID)
	assert.Equal(t, 2, item.Quantity)
	orderRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
}

func TestOrderService_AddToCart_ProductNotFound(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	productRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrProductNotFound)

	item, err := svc.AddToCart(ctx, 1, 999, 1)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, ErrProductNotFound)
	productRepo.AssertExpectations(t)
}

func TestOrderService_AddToCart_ProductUnavailable(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 10,
		Status:   "pending",
	}, nil)

	item, err := svc.AddToCart(ctx, 1, 1, 1)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, ErrProductUnavailable)
	productRepo.AssertExpectations(t)
}

func TestOrderService_AddToCart_InsufficientStock(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		Price:    100.0,
		Quantity: 5,
		Status:   "approved",
	}, nil)

	item, err := svc.AddToCart(ctx, 1, 1, 10)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, ErrInsufficientStock)
	productRepo.AssertExpectations(t)
}

func TestOrderService_GetCart_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetCartByUserID", ctx, int64(1)).Return([]*entity.CartItem{
		{ID: 1, UserID: 1, ProductID: 1, Quantity: 2},
		{ID: 2, UserID: 1, ProductID: 2, Quantity: 1},
	}, nil)

	productRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:    1,
		Price: 100.0,
	}, nil)
	productRepo.On("GetByID", ctx, int64(2)).Return(&entity.Product{
		ID:    2,
		Price: 50.0,
	}, nil)

	items, total, err := svc.GetCart(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 250.0, total)
	orderRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
}

func TestOrderService_GetCart_Empty(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("GetCartByUserID", ctx, int64(1)).Return([]*entity.CartItem{}, nil)

	items, total, err := svc.GetCart(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, items, 0)
	assert.Equal(t, 0.0, total)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateCartItem_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("UpdateCartItem", ctx, int64(1), 5).Return(nil)

	err := svc.UpdateCartItem(ctx, 1, 1, 5)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_RemoveFromCart_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("DeleteCartItem", ctx, int64(1)).Return(nil)

	err := svc.RemoveFromCart(ctx, 1, 1)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_ClearCart_Success(t *testing.T) {
	orderRepo := mocks.NewMockOrderRepository()
	productRepo := mocks.NewMockProductRepository()
	svc := newTestOrderService(orderRepo, productRepo)
	ctx := context.Background()

	orderRepo.On("ClearCart", ctx, int64(1)).Return(nil)

	err := svc.ClearCart(ctx, 1)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}
