package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"go.uber.org/zap"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) (*entity.Product, error)
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	GetAll(ctx context.Context, filter *dto.ProductFilter) ([]*entity.Product, int64, error)
	GetBySellerID(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) (*entity.Product, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateQuantity(ctx context.Context, id int64, quantity int) error
	Delete(ctx context.Context, id int64) error
}

type productRepository struct {
	txManager *db.TransactionManager
	qManager  *db.QueryManager
	logger    *zap.Logger
}

func NewProductRepository(txManager *db.TransactionManager, qManager *db.QueryManager, logger *zap.Logger) ProductRepository {
	return &productRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	r.logger.Info("Creating product", zap.String("title", product.Title))

	var createdProduct entity.Product
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			INSERT INTO products (seller_id, title, description, price, quantity, category, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, seller_id, title, description, price, quantity, category, status, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, product.SellerID, product.Title, product.Description,
			product.Price, product.Quantity, product.Category, product.Status)
		if err := row.Scan(&createdProduct.ID, &createdProduct.SellerID, &createdProduct.Title,
			&createdProduct.Description, &createdProduct.Price, &createdProduct.Quantity,
			&createdProduct.Category, &createdProduct.Status, &createdProduct.CreatedAt,
			&createdProduct.UpdatedAt); err != nil {
			r.logger.Error("Failed to create product", zap.Error(err))
			return fmt.Errorf("failed to create product: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.Info("Product created successfully", zap.Int64("id", createdProduct.ID))
	return &createdProduct, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	r.logger.Debug("Getting product by ID", zap.Int64("id", id))

	query := `
		SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	row := r.qManager.QueryRow(ctx, query, id)

	var product entity.Product
	err := row.Scan(&product.ID, &product.SellerID, &product.Title, &product.Description,
		&product.Price, &product.Quantity, &product.Category, &product.Status,
		&product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		r.logger.Error("Failed to get product", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context, filter *dto.ProductFilter) ([]*entity.Product, int64, error) {
	r.logger.Debug("Getting all products with filter")

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argNum))
		args = append(args, filter.Category)
		argNum++
	}

	if filter.MinPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argNum))
		args = append(args, filter.MinPrice)
		argNum++
	}

	if filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argNum))
		args = append(args, filter.MaxPrice)
		argNum++
	}

	if filter.SellerID > 0 {
		conditions = append(conditions, fmt.Sprintf("seller_id = $%d", argNum))
		args = append(args, filter.SellerID)
		argNum++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}

	whereCondition := ""
	if len(conditions) > 0 {
		whereCondition = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereCondition)
	row := r.qManager.QueryRow(ctx, countQuery, args...)
	var totalCount int64
	if err := row.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	limit := 20
	if filter.PageSize > 0 && filter.PageSize <= 100 {
		limit = filter.PageSize
	}
	offset := 0
	if filter.Page > 1 {
		offset = (filter.Page - 1) * limit
	}

	query := fmt.Sprintf(`
		SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at
		FROM products
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereCondition, sortBy, sortOrder, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.qManager.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var product entity.Product
		if err := rows.Scan(&product.ID, &product.SellerID, &product.Title, &product.Description,
			&product.Price, &product.Quantity, &product.Category, &product.Status,
			&product.CreatedAt, &product.UpdatedAt); err != nil {
			r.logger.Error("Failed to scan product", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, totalCount, nil
}

func (r *productRepository) GetBySellerID(ctx context.Context, sellerID int64, limit, offset int) ([]*entity.Product, error) {
	r.logger.Debug("Getting products by seller ID", zap.Int64("seller_id", sellerID))

	query := `
		SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at
		FROM products
		WHERE seller_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.qManager.Query(ctx, query, sellerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var product entity.Product
		if err := rows.Scan(&product.ID, &product.SellerID, &product.Title, &product.Description,
			&product.Price, &product.Quantity, &product.Category, &product.Status,
			&product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	r.logger.Info("Updating product", zap.Int64("id", product.ID))

	var updatedProduct entity.Product
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			UPDATE products
			SET title = $1, description = $2, price = $3, quantity = $4, category = $5
			WHERE id = $6
			RETURNING id, seller_id, title, description, price, quantity, category, status, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, product.Title, product.Description, product.Price,
			product.Quantity, product.Category, product.ID)
		if err := row.Scan(&updatedProduct.ID, &updatedProduct.SellerID, &updatedProduct.Title,
			&updatedProduct.Description, &updatedProduct.Price, &updatedProduct.Quantity,
			&updatedProduct.Category, &updatedProduct.Status, &updatedProduct.CreatedAt,
			&updatedProduct.UpdatedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrProductNotFound
			}
			r.logger.Error("Failed to update product", zap.Error(err))
			return fmt.Errorf("failed to update product: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &updatedProduct, nil
}

func (r *productRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	r.logger.Info("Updating product status", zap.Int64("id", id), zap.String("status", status))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE products SET status = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrProductNotFound
		}

		return nil
	})

	return err
}

func (r *productRepository) UpdateQuantity(ctx context.Context, id int64, quantity int) error {
	r.logger.Info("Updating product quantity", zap.Int64("id", id), zap.Int("quantity", quantity))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE products SET quantity = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, quantity, id)
		if err != nil {
			return fmt.Errorf("failed to update quantity: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrProductNotFound
		}

		return nil
	})

	return err
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	r.logger.Info("Deleting product", zap.Int64("id", id))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `DELETE FROM products WHERE id = $1`
		result, err := tx.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete product: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrProductNotFound
		}

		return nil
	})

	return err
}
