package service

import (
	"context"
	"testing"
	"time"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/repository"
	"github.com/Thorlik/marketplace/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestProductService(mockRepo *mocks.MockProductRepository) ProductService {
	logger := zap.NewNop()
	return NewProductService(mockRepo, logger)
}

func TestProductService_Create_Success(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	product := &domain.Product{
		Title:       "Test Product",
		Description: "Test Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.Product")).Return(&entity.Product{
		ID:          1,
		SellerID:    1,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	result, err := svc.Create(ctx, 1, product)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Product", result.Title)
	assert.Equal(t, domain.ProductStatusPending, result.Status)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Create_InvalidTitle_Empty(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	product := &domain.Product{
		Title:    "",
		Price:    100.0,
		Quantity: 10,
	}

	result, err := svc.Create(ctx, 1, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidProductTitle)
}

func TestProductService_Create_InvalidTitle_TooLong(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	longTitle := ""
	for i := 0; i < 256; i++ {
		longTitle += "a"
	}

	product := &domain.Product{
		Title:    longTitle,
		Price:    100.0,
		Quantity: 10,
	}

	result, err := svc.Create(ctx, 1, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidProductTitle)
}

func TestProductService_Create_InvalidPrice_Zero(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	product := &domain.Product{
		Title:    "Test Product",
		Price:    0,
		Quantity: 10,
	}

	result, err := svc.Create(ctx, 1, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidProductPrice)
}

func TestProductService_Create_InvalidPrice_Negative(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	product := &domain.Product{
		Title:    "Test Product",
		Price:    -100.0,
		Quantity: 10,
	}

	result, err := svc.Create(ctx, 1, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidProductPrice)
}

func TestProductService_GetByID_Success(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:          1,
		SellerID:    1,
		Title:       "Test Product",
		Description: "Description",
		Price:       100.0,
		Quantity:    10,
		Category:    "Electronics",
		Status:      "approved",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	result, err := svc.GetByID(ctx, 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetByID_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrProductNotFound)

	result, err := svc.GetByID(ctx, 999)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrProductNotFound)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetAll_Success(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	filter := &dto.ProductFilter{
		Page:     1,
		PageSize: 20,
	}

	mockRepo.On("GetAll", ctx, filter).Return([]*entity.Product{
		{ID: 1, Title: "Product 1", Price: 100, Status: "approved", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, Title: "Product 2", Price: 200, Status: "approved", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, int64(2), nil)

	products, total, err := svc.GetAll(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetAll_InvalidPagination(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	filter := &dto.ProductFilter{
		Page:     -1,
		PageSize: 200,
	}

	mockRepo.On("GetAll", ctx, mock.MatchedBy(func(f *dto.ProductFilter) bool {
		return f.Page == 1 && f.PageSize == 20
	})).Return([]*entity.Product{}, int64(0), nil)

	_, _, err := svc.GetAll(ctx, filter)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetBySeller_Success(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetBySellerID", ctx, int64(1), 20, 0).Return([]*entity.Product{
		{ID: 1, SellerID: 1, Title: "My Product", Price: 100, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}, nil)

	products, err := svc.GetBySeller(ctx, 1, 1, 20)

	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, int64(1), products[0].SellerID)
	mockRepo.AssertExpectations(t)
}

func TestProductService_GetBySeller_InvalidPagination(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetBySellerID", ctx, int64(1), 20, 0).Return([]*entity.Product{}, nil)

	_, err := svc.GetBySeller(ctx, 1, 0, 200)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Update_Success_Owner(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	existingProduct := &entity.Product{
		ID:       1,
		SellerID: 1,
		Title:    "Old Title",
		Price:    100,
		Status:   "pending",
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(existingProduct, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(&entity.Product{
		ID:        1,
		SellerID:  1,
		Title:     "New Title",
		Price:     150,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	product := &domain.Product{
		ID:    1,
		Title: "New Title",
		Price: 150,
	}

	result, err := svc.Update(ctx, 1, domain.RoleUser, product)

	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Update_Success_Admin(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	existingProduct := &entity.Product{
		ID:       1,
		SellerID: 2,
		Title:    "Old Title",
		Price:    100,
		Status:   "pending",
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(existingProduct, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(&entity.Product{
		ID:        1,
		SellerID:  2,
		Title:     "Updated by Admin",
		Price:     200,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	product := &domain.Product{
		ID:    1,
		Title: "Updated by Admin",
		Price: 200,
	}

	result, err := svc.Update(ctx, 999, domain.RoleAdmin, product)
	require.NoError(t, err)
	assert.Equal(t, "Updated by Admin", result.Title)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Update_NotOwner(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	existingProduct := &entity.Product{
		ID:       1,
		SellerID: 2,
		Title:    "Product",
		Price:    100,
	}

	mockRepo.On("GetByID", ctx, int64(1)).Return(existingProduct, nil)

	product := &domain.Product{
		ID:    1,
		Title: "Hacked",
		Price: 200,
	}

	result, err := svc.Update(ctx, 1, domain.RoleUser, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotProductOwner)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Update_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrProductNotFound)

	product := &domain.Product{
		ID:    999,
		Title: "Test",
		Price: 100,
	}

	result, err := svc.Update(ctx, 1, domain.RoleUser, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrProductNotFound)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Update_InvalidTitle(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	product := &domain.Product{
		ID:    1,
		Title: "",
		Price: 100,
	}

	result, err := svc.Update(ctx, 1, domain.RoleUser, product)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidProductTitle)
}

func TestProductService_UpdateStatus_Success(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateStatus", ctx, int64(1), "approved").Return(nil)

	err := svc.UpdateStatus(ctx, 1, domain.ProductStatusApproved)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_UpdateStatus_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateStatus", ctx, int64(999), "approved").Return(repository.ErrProductNotFound)

	err := svc.UpdateStatus(ctx, 999, domain.ProductStatusApproved)

	assert.ErrorIs(t, err, ErrProductNotFound)
	mockRepo.AssertExpectations(t)
}

func TestProductService_UpdateStatus_ToRejected(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("UpdateStatus", ctx, int64(1), "rejected").Return(nil)

	err := svc.UpdateStatus(ctx, 1, domain.ProductStatusRejected)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_Success_Owner(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		SellerID: 1,
	}, nil)
	mockRepo.On("Delete", ctx, int64(1)).Return(nil)

	err := svc.Delete(ctx, 1, domain.RoleUser, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_Success_Admin(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		SellerID: 2,
	}, nil)
	mockRepo.On("Delete", ctx, int64(1)).Return(nil)

	err := svc.Delete(ctx, 999, domain.RoleAdmin, 1)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_NotOwner(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		SellerID: 2,
	}, nil)

	err := svc.Delete(ctx, 1, domain.RoleUser, 1)

	assert.ErrorIs(t, err, ErrNotProductOwner)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_NotOwner_Moderator(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(1)).Return(&entity.Product{
		ID:       1,
		SellerID: 2,
	}, nil)

	err := svc.Delete(ctx, 1, domain.RoleModerator, 1)

	assert.ErrorIs(t, err, ErrNotProductOwner)
	mockRepo.AssertExpectations(t)
}

func TestProductService_Delete_NotFound(t *testing.T) {
	mockRepo := mocks.NewMockProductRepository()
	svc := newTestProductService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(999)).Return(nil, repository.ErrProductNotFound)

	err := svc.Delete(ctx, 1, domain.RoleUser, 999)

	assert.ErrorIs(t, err, ErrProductNotFound)
	mockRepo.AssertExpectations(t)
}
