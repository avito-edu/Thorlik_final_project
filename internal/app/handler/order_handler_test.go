package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/Thorlik/marketplace/internal/service/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestOrderHandler(mockService *mocks.MockOrderService) *OrderHandler {
	logger := zap.NewNop()
	return NewOrderHandler(mockService, logger)
}

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("CreateOrder", mock.Anything, int64(1), mock.AnythingOfType("[]domain.OrderItem")).Return(&domain.Order{
		ID:            1,
		UserID:        1,
		Status:        domain.OrderStatusPending,
		TotalAmount:   250.0,
		PaymentStatus: domain.PaymentStatusPending,
		CreatedAt:     time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrder(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response dto.OrderResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, 250.0, response.TotalAmount)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_InvalidBody(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrder(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_CreateOrder_EmptyCart(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateOrderRequest{Items: []dto.OrderItemRequest{}}
	body, _ := json.Marshal(reqBody)

	mockService.On("CreateOrder", mock.Anything, int64(1), mock.AnythingOfType("[]domain.OrderItem")).Return(nil, service.ErrCartEmpty)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrder(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_InsufficientStock(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateOrderRequest{
		Items: []dto.OrderItemRequest{
			{ProductID: 1, Quantity: 100},
		},
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("CreateOrder", mock.Anything, int64(1), mock.AnythingOfType("[]domain.OrderItem")).Return(nil, service.ErrInsufficientStock)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrder(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrderFromCart_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("CreateOrderFromCart", mock.Anything, int64(1)).Return(&domain.Order{
		ID:            1,
		UserID:        1,
		Status:        domain.OrderStatusPending,
		TotalAmount:   200.0,
		PaymentStatus: domain.PaymentStatusPending,
		CreatedAt:     time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/from-cart", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrderFromCart(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_CreateOrderFromCart_EmptyCart(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("CreateOrderFromCart", mock.Anything, int64(1)).Return(nil, service.ErrCartEmpty)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/from-cart", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateOrderFromCart(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("GetOrderByID", mock.Anything, int64(1), domain.RoleUser, int64(1)).Return(&domain.Order{
		ID:            1,
		UserID:        1,
		Status:        domain.OrderStatusPending,
		TotalAmount:   100.0,
		PaymentStatus: domain.PaymentStatusPending,
		CreatedAt:     time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_InvalidID(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	req := httptest.NewRequest(http.MethodGet, "/api/orders/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("GetOrderByID", mock.Anything, int64(1), domain.RoleUser, int64(999)).Return(nil, service.ErrOrderNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_NotOwner(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 2, Email: "other@example.com", Role: domain.RoleUser}

	mockService.On("GetOrderByID", mock.Anything, int64(2), domain.RoleUser, int64(1)).Return(nil, service.ErrNotOrderOwner)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetOrder(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetMyOrders_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("GetUserOrders", mock.Anything, int64(1), 1, 20).Return([]*domain.Order{
		{ID: 1, UserID: 1, Status: domain.OrderStatusPending, TotalAmount: 100, CreatedAt: time.Now()},
		{ID: 2, UserID: 1, Status: domain.OrderStatusPaid, TotalAmount: 200, CreatedAt: time.Now()},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/orders?page=1&page_size=20", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetMyOrders(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []dto.OrderResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 2)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetAllOrders_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	mockService.On("GetAllOrders", mock.Anything, 1, 20).Return([]*domain.Order{
		{ID: 1, UserID: 1, Status: domain.OrderStatusPending, TotalAmount: 100, CreatedAt: time.Now()},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orders?page=1&page_size=20", nil)
	rr := httptest.NewRecorder()

	handler.GetAllOrders(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_UpdateOrderStatus_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	reqBody := dto.UpdateOrderStatusRequest{Status: "shipped"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateOrderStatus", mock.Anything, int64(1), domain.OrderStatusShipped).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/1/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.UpdateOrderStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_UpdateOrderStatus_InvalidID(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	reqBody := dto.UpdateOrderStatusRequest{Status: "shipped"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/invalid/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rr := httptest.NewRecorder()

	handler.UpdateOrderStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateOrderStatus_NotFound(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	reqBody := dto.UpdateOrderStatusRequest{Status: "shipped"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateOrderStatus", mock.Anything, int64(999), domain.OrderStatusShipped).Return(service.ErrOrderNotFound)

	req := httptest.NewRequest(http.MethodPut, "/api/orders/999/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.UpdateOrderStatus(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_ProcessPayment_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.PaymentRequest{
		OrderID:       1,
		PaymentMethod: "card",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("ProcessPayment", mock.Anything, int64(1), int64(1), "card").Return(&service.PaymentResult{
		PaymentID:     "PAY-1-1",
		Status:        "success",
		Message:       "Payment processed",
		TransactionID: "TXN-1",
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/pay", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ProcessPayment(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.PaymentResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_ProcessPayment_NotOwner(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 2, Email: "other@example.com", Role: domain.RoleUser}

	reqBody := dto.PaymentRequest{
		OrderID:       1,
		PaymentMethod: "card",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("ProcessPayment", mock.Anything, int64(2), int64(1), "card").Return(nil, service.ErrNotOrderOwner)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/pay", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ProcessPayment(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_AddToCart_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.AddToCartRequest{
		ProductID: 1,
		Quantity:  2,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("AddToCart", mock.Anything, int64(1), int64(1), 2).Return(&domain.CartItem{
		ID:        1,
		UserID:    1,
		ProductID: 1,
		Quantity:  2,
		CreatedAt: time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddToCart(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.CartItemResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ProductID)
	assert.Equal(t, 2, response.Quantity)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_AddToCart_ProductNotFound(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.AddToCartRequest{
		ProductID: 999,
		Quantity:  1,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("AddToCart", mock.Anything, int64(1), int64(999), 1).Return(nil, service.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddToCart(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_AddToCart_InsufficientStock(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.AddToCartRequest{
		ProductID: 1,
		Quantity:  100,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("AddToCart", mock.Anything, int64(1), int64(1), 100).Return(nil, service.ErrInsufficientStock)

	req := httptest.NewRequest(http.MethodPost, "/api/cart", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.AddToCart(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetCart_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("GetCart", mock.Anything, int64(1)).Return([]*domain.CartItem{
		{ID: 1, UserID: 1, ProductID: 1, Quantity: 2, Product: &domain.Product{ID: 1, Title: "Product", Price: 100}},
	}, 200.0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetCart(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.CartResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Items, 1)
	assert.Equal(t, 200.0, response.TotalPrice)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_GetCart_Empty(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("GetCart", mock.Anything, int64(1)).Return([]*domain.CartItem{}, 0.0, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetCart(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_UpdateCartItem_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateCartItemRequest{Quantity: 5}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateCartItem", mock.Anything, int64(1), int64(1), 5).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/cart/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateCartItem(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_UpdateCartItem_InvalidID(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateCartItemRequest{Quantity: 5}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/cart/invalid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateCartItem(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_RemoveFromCart_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("RemoveFromCart", mock.Anything, int64(1), int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/cart/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.RemoveFromCart(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestOrderHandler_RemoveFromCart_InvalidID(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	req := httptest.NewRequest(http.MethodDelete, "/api/cart/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.RemoveFromCart(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_ClearCart_Success(t *testing.T) {
	mockService := mocks.NewMockOrderService()
	handler := newTestOrderHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "user@example.com", Role: domain.RoleUser}

	mockService.On("ClearCart", mock.Anything, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/cart/clear", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ClearCart(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}
