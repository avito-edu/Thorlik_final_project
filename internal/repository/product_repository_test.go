package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupProductRepositoryTest(t *testing.T) (*productRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	txManager := db.NewTransactionManager(mockDB, logger)
	qManager := db.NewQueryManager(mockDB, logger)

	repo := &productRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestProductRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		SellerID:    1,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Quantity:    10,
		Category:    "Electronics",
		Status:      "pending",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, product.SellerID, product.Title, product.Description, product.Price, product.Quantity, product.Category, product.Status, now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs(product.SellerID, product.Title, product.Description, product.Price, product.Quantity, product.Category, product.Status).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.Create(context.Background(), product)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, product.Title, result.Title)
	assert.Equal(t, product.Price, result.Price)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Create_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		SellerID:    1,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Quantity:    10,
		Category:    "Electronics",
		Status:      "pending",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs(product.SellerID, product.Title, product.Description, product.Price, product.Quantity, product.Category, product.Status).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	result, err := repo.Create(context.Background(), product)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Create_BeginTransactionError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		SellerID: 1,
		Title:    "Test Product",
	}

	mock.ExpectBegin().WillReturnError(errors.New("connection error"))

	result, err := repo.Create(context.Background(), product)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Test Product", "Description", 99.99, 10, "Electronics", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Test Product", result.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrProductNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetByID(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products`).
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Product 1", "Desc 1", 99.99, 10, "Electronics", "approved", now, now).
		AddRow(2, 2, "Product 2", "Desc 2", 49.99, 5, "Books", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "Product 1", result[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_WithCategoryFilter(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Category: "Electronics",
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE category = \$1`).
		WithArgs("Electronics").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Electronics Product", "Desc", 199.99, 5, "Electronics", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE category = \$1`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Electronics", result[0].Category)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_Empty(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Page:     1,
		PageSize: 10,
	}

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products`).
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"})
	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_CountError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Page:     1,
		PageSize: 10,
	}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products`).
		WillReturnError(errors.New("database error"))

	result, total, err := repo.GetAll(context.Background(), filter)

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_QueryError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Page:     1,
		PageSize: 10,
	}

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products`).
		WillReturnRows(countRows)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products`).
		WillReturnError(errors.New("database error"))

	result, total, err := repo.GetAll(context.Background(), filter)

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetBySellerID_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Product 1", "Desc 1", 99.99, 10, "Electronics", "approved", now, now).
		AddRow(2, 1, "Product 2", "Desc 2", 49.99, 5, "Electronics", "pending", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE seller_id = \$1`).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	result, err := repo.GetBySellerID(context.Background(), 1, 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].SellerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetBySellerID_Empty(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE seller_id = \$1`).
		WithArgs(int64(999), 10, 0).
		WillReturnRows(rows)

	result, err := repo.GetBySellerID(context.Background(), 999, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetBySellerID_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE seller_id = \$1`).
		WithArgs(int64(1), 10, 0).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetBySellerID(context.Background(), 1, 10, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		ID:          1,
		Title:       "Updated Product",
		Description: "Updated Description",
		Price:       149.99,
		Quantity:    15,
		Category:    "Electronics",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(product.ID, 1, product.Title, product.Description, product.Price, product.Quantity, product.Category, "approved", now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE products SET title = \$1, description = \$2, price = \$3, quantity = \$4, category = \$5 WHERE id = \$6`).
		WithArgs(product.Title, product.Description, product.Price, product.Quantity, product.Category, product.ID).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.Update(context.Background(), product)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, product.Title, result.Title)
	assert.Equal(t, product.Price, result.Price)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		ID:          999,
		Title:       "Updated Product",
		Description: "Updated Description",
		Price:       149.99,
		Quantity:    15,
		Category:    "Electronics",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE products SET title = \$1, description = \$2, price = \$3, quantity = \$4, category = \$5 WHERE id = \$6`).
		WithArgs(product.Title, product.Description, product.Price, product.Quantity, product.Category, product.ID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	result, err := repo.Update(context.Background(), product)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrProductNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		ID:          1,
		Title:       "Updated Product",
		Description: "Updated Description",
		Price:       149.99,
		Quantity:    15,
		Category:    "Electronics",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE products SET title = \$1, description = \$2, price = \$3, quantity = \$4, category = \$5 WHERE id = \$6`).
		WithArgs(product.Title, product.Description, product.Price, product.Quantity, product.Category, product.ID).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	result, err := repo.Update(context.Background(), product)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateStatus_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("approved", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(context.Background(), 1, "approved")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateStatus_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("approved", int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateStatus(context.Background(), 999, "approved")

	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateStatus_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET status = \$1 WHERE id = \$2`).
		WithArgs("approved", int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdateStatus(context.Background(), 1, "approved")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateQuantity_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET quantity = \$1 WHERE id = \$2`).
		WithArgs(20, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateQuantity(context.Background(), 1, 20)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateQuantity_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET quantity = \$1 WHERE id = \$2`).
		WithArgs(20, int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateQuantity(context.Background(), 999, 20)

	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_UpdateQuantity_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE products SET quantity = \$1 WHERE id = \$2`).
		WithArgs(20, int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdateQuantity(context.Background(), 1, 20)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_Success(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_NotFound(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, ErrProductNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_DBError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Create_CommitError(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	product := &entity.Product{
		SellerID:    1,
		Title:       "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Quantity:    10,
		Category:    "Electronics",
		Status:      "pending",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, product.SellerID, product.Title, product.Description, product.Price, product.Quantity, product.Category, product.Status, now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO products`).
		WithArgs(product.SellerID, product.Title, product.Description, product.Price, product.Quantity, product.Category, product.Status).
		WillReturnRows(rows)
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	result, err := repo.Create(context.Background(), product)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_WithPriceFilter(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		MinPrice: 50,
		MaxPrice: 200,
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE price >= \$1 AND price <= \$2`).
		WithArgs(float64(50), float64(200)).
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Product", "Desc", 99.99, 10, "Electronics", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE price >= \$1 AND price <= \$2`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_WithSellerFilter(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		SellerID: 1,
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE seller_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Product 1", "Desc 1", 99.99, 10, "Electronics", "approved", now, now).
		AddRow(2, 1, "Product 2", "Desc 2", 49.99, 5, "Books", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE seller_id = \$1`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_WithStatusFilter(t *testing.T) {
	repo, mock, cleanup := setupProductRepositoryTest(t)
	defer cleanup()

	filter := &dto.ProductFilter{
		Status:   "approved",
		Page:     1,
		PageSize: 10,
	}

	now := time.Now()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM products WHERE status = \$1`).
		WithArgs("approved").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price", "quantity", "category", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "Product", "Desc", 99.99, 10, "Electronics", "approved", now, now)

	mock.ExpectQuery(`SELECT id, seller_id, title, description, price, quantity, category, status, created_at, updated_at FROM products WHERE status = \$1`).
		WillReturnRows(rows)

	result, total, err := repo.GetAll(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "approved", result[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}
