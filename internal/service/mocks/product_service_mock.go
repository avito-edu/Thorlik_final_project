package mocks

import (
	"context"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/stretchr/testify/mock"
)

type MockProductService struct {
	mock.Mock
}

func NewMockProductService() *MockProductService {
	return &MockProductService{}
}

func (m *MockProductService) Create(ctx context.Context, sellerID int64, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, sellerID, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) GetAll(ctx context.Context, filter *dto.ProductFilter) ([]*domain.Product, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductService) GetBySeller(ctx context.Context, sellerID int64, page, pageSize int) ([]*domain.Product, error) {
	args := m.Called(ctx, sellerID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *MockProductService) Update(ctx context.Context, userID int64, userRole domain.UserRole, product *domain.Product) (*domain.Product, error) {
	args := m.Called(ctx, userID, userRole, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductService) UpdateStatus(ctx context.Context, id int64, status domain.ProductStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockProductService) Delete(ctx context.Context, userID int64, userRole domain.UserRole, productID int64) error {
	args := m.Called(ctx, userID, userRole, productID)
	return args.Error(0)
}
