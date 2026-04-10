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

func newTestProductHandler(mockService *mocks.MockProductService) *ProductHandler {
	logger := zap.NewNop()
	return NewProductHandler(mockService, logger)
}

func TestProductHandler_CreateProduct_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateProductRequest{
		Title:       "Test Product",
		Description: "Test Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Create", mock.Anything, int64(1), mock.AnythingOfType("*domain.Product")).Return(&domain.Product{
		ID:          1,
		SellerID:    1,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
		Status:      domain.ProductStatusPending,
		CreatedAt:   time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateProduct(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response dto.ProductResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, "Test Product", response.Title)
	mockService.AssertExpectations(t)
}

func TestProductHandler_CreateProduct_InvalidBody(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProductHandler_CreateProduct_InvalidTitle(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateProductRequest{
		Title: "",
		Price: 100.0,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Create", mock.Anything, int64(1), mock.AnythingOfType("*domain.Product")).Return(nil, service.ErrInvalidProductTitle)

	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_CreateProduct_InvalidPrice(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	reqBody := dto.CreateProductRequest{
		Title: "Product",
		Price: -10.0,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Create", mock.Anything, int64(1), mock.AnythingOfType("*domain.Product")).Return(nil, service.ErrInvalidProductPrice)

	req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.CreateProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	mockService.On("GetByID", mock.Anything, int64(1)).Return(&domain.Product{
		ID:          1,
		SellerID:    1,
		Title:       "Test Product",
		Description: "Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
		Status:      domain.ProductStatusApproved,
		CreatedAt:   time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.GetProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.ProductResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetProduct_InvalidID(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/products/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rr := httptest.NewRecorder()

	handler.GetProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProductHandler_GetProduct_NotFound(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, service.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/products/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.GetProduct(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetAllProducts_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	mockService.On("GetAll", mock.Anything, mock.AnythingOfType("*dto.ProductFilter")).Return([]*domain.Product{
		{ID: 1, Title: "Product 1", Price: 100, Status: domain.ProductStatusApproved, CreatedAt: time.Now()},
		{ID: 2, Title: "Product 2", Price: 200, Status: domain.ProductStatusApproved, CreatedAt: time.Now()},
	}, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/products?page=1&page_size=20", nil)
	rr := httptest.NewRecorder()

	handler.GetAllProducts(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.ProductListResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Products, 2)
	assert.Equal(t, int64(2), response.TotalCount)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetMyProducts_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	mockService.On("GetBySeller", mock.Anything, int64(1), 1, 20).Return([]*domain.Product{
		{ID: 1, SellerID: 1, Title: "My Product", Price: 100, CreatedAt: time.Now()},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/products/my?page=1&page_size=20", nil)
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetMyProducts(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []dto.ProductResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response, 1)
	mockService.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateProductRequest{
		Title:       "Updated Product",
		Description: "Updated Description",
		Price:       150.0,
		Quantity:    5,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Update", mock.Anything, int64(1), domain.RoleUser, mock.AnythingOfType("*domain.Product")).Return(&domain.Product{
		ID:          1,
		SellerID:    1,
		Title:       "Updated Product",
		Description: "Updated Description",
		Price:       150.0,
		Quantity:    5,
		Status:      domain.ProductStatusPending,
		CreatedAt:   time.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.ProductResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Product", response.Title)
	mockService.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_NotOwner(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 2, Email: "other@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateProductRequest{
		Title: "Hacked",
		Price: 100.0,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Update", mock.Anything, int64(2), domain.RoleUser, mock.AnythingOfType("*domain.Product")).Return(nil, service.ErrNotProductOwner)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateProduct(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_NotFound(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	reqBody := dto.UpdateProductRequest{
		Title: "Product",
		Price: 100.0,
	}
	body, _ := json.Marshal(reqBody)

	mockService.On("Update", mock.Anything, int64(1), domain.RoleUser, mock.AnythingOfType("*domain.Product")).Return(nil, service.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodPut, "/api/products/999", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateProduct(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	mockService.On("Delete", mock.Anything, int64(1), domain.RoleUser, int64(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.DeleteProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_NotOwner(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 2, Email: "other@example.com", Role: domain.RoleUser}

	mockService.On("Delete", mock.Anything, int64(2), domain.RoleUser, int64(1)).Return(service.ErrNotProductOwner)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.DeleteProduct(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_NotFound(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	claims := &service.Claims{UserID: 1, Email: "seller@example.com", Role: domain.RoleUser}

	mockService.On("Delete", mock.Anything, int64(1), domain.RoleUser, int64(999)).Return(service.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/products/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	ctx := context.WithValue(req.Context(), "claims", claims)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.DeleteProduct(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_ModerateProduct_Approve_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	reqBody := dto.ModerateProductRequest{Status: "approved"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateStatus", mock.Anything, int64(1), domain.ProductStatusApproved).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1/moderate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ModerateProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_ModerateProduct_Reject_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	reqBody := dto.ModerateProductRequest{Status: "rejected"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateStatus", mock.Anything, int64(1), domain.ProductStatusRejected).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1/moderate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ModerateProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_ModerateProduct_InvalidStatus(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	reqBody := dto.ModerateProductRequest{Status: "invalid"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/products/1/moderate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	handler.ModerateProduct(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestProductHandler_ModerateProduct_NotFound(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	reqBody := dto.ModerateProductRequest{Status: "approved"}
	body, _ := json.Marshal(reqBody)

	mockService.On("UpdateStatus", mock.Anything, int64(999), domain.ProductStatusApproved).Return(service.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodPut, "/api/products/999/moderate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	rr := httptest.NewRecorder()

	handler.ModerateProduct(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestProductHandler_GetPendingProducts_Success(t *testing.T) {
	mockService := mocks.NewMockProductService()
	handler := newTestProductHandler(mockService)

	mockService.On("GetAll", mock.Anything, mock.MatchedBy(func(f *dto.ProductFilter) bool {
		return f.Status == "pending"
	})).Return([]*domain.Product{
		{ID: 1, Title: "Pending Product", Price: 100, Status: domain.ProductStatusPending, CreatedAt: time.Now()},
	}, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/moderation/products", nil)
	rr := httptest.NewRecorder()

	handler.GetPendingProducts(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.ProductListResponse
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)
	assert.Len(t, response.Products, 1)
	mockService.AssertExpectations(t)
}
