package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidProductTitle = errors.New("product title must be between 1 and 255 characters")
	ErrInvalidProductPrice = errors.New("product price must be greater than 0")
	ErrNotProductOwner     = errors.New("you are not the owner of this product")
)

type ProductService interface {
	Create(ctx context.Context, sellerID int64, product *domain.Product) (*domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	GetAll(ctx context.Context, filter *dto.ProductFilter) ([]*domain.Product, int64, error)
	GetBySeller(ctx context.Context, sellerID int64, page, pageSize int) ([]*domain.Product, error)
	Update(ctx context.Context, userID int64, userRole domain.UserRole, product *domain.Product) (*domain.Product, error)
	UpdateStatus(ctx context.Context, id int64, status domain.ProductStatus) error
	Delete(ctx context.Context, userID int64, userRole domain.UserRole, productID int64) error
}

type productService struct {
	repo   repository.ProductRepository
	logger *zap.Logger
}

func NewProductService(repo repository.ProductRepository, logger *zap.Logger) ProductService {
	return &productService{
		repo:   repo,
		logger: logger,
	}
}

func (s *productService) validateProduct(product *domain.Product) error {
	if len(product.Title) < 1 || len(product.Title) > 255 {
		return ErrInvalidProductTitle
	}
	if product.Price <= 0 {
		return ErrInvalidProductPrice
	}
	return nil
}

func (s *productService) Create(ctx context.Context, sellerID int64, product *domain.Product) (*domain.Product, error) {
	s.logger.Info("Creating product", zap.String("title", product.Title), zap.Int64("seller_id", sellerID))

	if err := s.validateProduct(product); err != nil {
		return nil, err
	}

	product.SellerID = sellerID
	product.Status = domain.ProductStatusPending

	entityProduct := model.ProductDomainToEntity(product)
	createdProduct, err := s.repo.Create(ctx, entityProduct)
	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	s.logger.Info("Product created successfully", zap.Int64("id", createdProduct.ID))
	return model.ProductEntityToDomain(createdProduct), nil
}

func (s *productService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	entityProduct, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return model.ProductEntityToDomain(entityProduct), nil
}

func (s *productService) GetAll(ctx context.Context, filter *dto.ProductFilter) ([]*domain.Product, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	entityProducts, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}

	products := make([]*domain.Product, 0, len(entityProducts))
	for _, ep := range entityProducts {
		products = append(products, model.ProductEntityToDomain(ep))
	}

	return products, totalCount, nil
}

func (s *productService) GetBySeller(ctx context.Context, sellerID int64, page, pageSize int) ([]*domain.Product, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	entityProducts, err := s.repo.GetBySellerID(ctx, sellerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	products := make([]*domain.Product, 0, len(entityProducts))
	for _, ep := range entityProducts {
		products = append(products, model.ProductEntityToDomain(ep))
	}

	return products, nil
}

func (s *productService) Update(ctx context.Context, userID int64, userRole domain.UserRole, product *domain.Product) (*domain.Product, error) {
	s.logger.Info("Updating product", zap.Int64("id", product.ID))

	if err := s.validateProduct(product); err != nil {
		return nil, err
	}

	existingProduct, err := s.repo.GetByID(ctx, product.ID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if existingProduct.SellerID != userID && userRole != domain.RoleAdmin && userRole != domain.RoleModerator {
		return nil, ErrNotProductOwner
	}

	entityProduct := model.ProductDomainToEntity(product)
	entityProduct.SellerID = existingProduct.SellerID
	entityProduct.Status = existingProduct.Status

	updatedProduct, err := s.repo.Update(ctx, entityProduct)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return model.ProductEntityToDomain(updatedProduct), nil
}

func (s *productService) UpdateStatus(ctx context.Context, id int64, status domain.ProductStatus) error {
	s.logger.Info("Updating product status", zap.Int64("id", id), zap.String("status", string(status)))

	if err := s.repo.UpdateStatus(ctx, id, string(status)); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (s *productService) Delete(ctx context.Context, userID int64, userRole domain.UserRole, productID int64) error {
	s.logger.Info("Deleting product", zap.Int64("id", productID))

	existingProduct, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to get product: %w", err)
	}

	if existingProduct.SellerID != userID && userRole != domain.RoleAdmin {
		return ErrNotProductOwner
	}

	if err := s.repo.Delete(ctx, productID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
