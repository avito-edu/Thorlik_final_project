package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupOrderRepositoryTest(t *testing.T) (*orderRepository, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop()
	txManager := db.NewTransactionManager(mockDB, logger)
	qManager := db.NewQueryManager(mockDB, logger)

	repo := &orderRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}

	cleanup := func() {
		mockDB.Close()
	}

	return repo, mock, cleanup
}

func TestOrderRepository_CreateOrder_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	order := &entity.Order{
		UserID:        1,
		Status:        "pending",
		TotalAmount:   199.99,
		PaymentStatus: "pending",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, order.UserID, order.Status, order.TotalAmount, order.PaymentStatus, nil, now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(order.UserID, order.Status, order.TotalAmount, order.PaymentStatus).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.CreateOrder(context.Background(), order)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, order.TotalAmount, result.TotalAmount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrder_WithPaymentID(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	order := &entity.Order{
		UserID:        1,
		Status:        "pending",
		TotalAmount:   199.99,
		PaymentStatus: "completed",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, order.UserID, order.Status, order.TotalAmount, order.PaymentStatus, "pay_123", now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(order.UserID, order.Status, order.TotalAmount, order.PaymentStatus).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.CreateOrder(context.Background(), order)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "pay_123", result.PaymentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrder_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	order := &entity.Order{
		UserID:        1,
		Status:        "pending",
		TotalAmount:   199.99,
		PaymentStatus: "pending",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(order.UserID, order.Status, order.TotalAmount, order.PaymentStatus).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	result, err := repo.CreateOrder(context.Background(), order)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrder_BeginTransactionError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	order := &entity.Order{
		UserID:        1,
		Status:        "pending",
		TotalAmount:   199.99,
		PaymentStatus: "pending",
	}

	mock.ExpectBegin().WillReturnError(errors.New("connection error"))

	result, err := repo.CreateOrder(context.Background(), order)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderByID_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, 1, "pending", 199.99, "pending", nil, now, now)

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetOrderByID(context.Background(), 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, 199.99, result.TotalAmount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderByID_WithPaymentID(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, 1, "completed", 199.99, "completed", "pay_123", now, now)

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetOrderByID(context.Background(), 1)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "pay_123", result.PaymentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetOrderByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrOrderNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderByID_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetOrderByID(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrdersByUserID_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, 1, "completed", 199.99, "completed", "pay_1", now, now).
		AddRow(2, 1, "pending", 99.99, "pending", nil, now, now)

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE user_id = \$1`).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	result, err := repo.GetOrdersByUserID(context.Background(), 1, 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].UserID)
	assert.Equal(t, "pay_1", result[0].PaymentID)
	assert.Empty(t, result[1].PaymentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrdersByUserID_Empty(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE user_id = \$1`).
		WithArgs(int64(999), 10, 0).
		WillReturnRows(rows)

	result, err := repo.GetOrdersByUserID(context.Background(), 999, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrdersByUserID_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders WHERE user_id = \$1`).
		WithArgs(int64(1), 10, 0).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetOrdersByUserID(context.Background(), 1, 10, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetAllOrders_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, 1, "completed", 199.99, "completed", "pay_1", now, now).
		AddRow(2, 2, "pending", 99.99, "pending", nil, now, now)

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.GetAllOrders(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetAllOrders_Empty(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.GetAllOrders(context.Background(), 10, 0)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetAllOrders_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetAllOrders(context.Background(), 10, 0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateOrderStatus_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$1 WHERE id = \$2`).
		WithArgs("shipped", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateOrderStatus(context.Background(), 1, "shipped")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateOrderStatus_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$1 WHERE id = \$2`).
		WithArgs("shipped", int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateOrderStatus(context.Background(), 999, "shipped")

	assert.Error(t, err)
	assert.Equal(t, ErrOrderNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateOrderStatus_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET status = \$1 WHERE id = \$2`).
		WithArgs("shipped", int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdateOrderStatus(context.Background(), 1, "shipped")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdatePaymentStatus_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET payment_status = \$1, payment_id = \$2 WHERE id = \$3`).
		WithArgs("completed", "pay_123", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdatePaymentStatus(context.Background(), 1, "completed", "pay_123")

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdatePaymentStatus_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET payment_status = \$1, payment_id = \$2 WHERE id = \$3`).
		WithArgs("completed", "pay_123", int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdatePaymentStatus(context.Background(), 999, "completed", "pay_123")

	assert.Error(t, err)
	assert.Equal(t, ErrOrderNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdatePaymentStatus_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE orders SET payment_status = \$1, payment_id = \$2 WHERE id = \$3`).
		WithArgs("completed", "pay_123", int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdatePaymentStatus(context.Background(), 1, "completed", "pay_123")

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrderItem_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	item := &entity.OrderItem{
		OrderID:         1,
		ProductID:       2,
		Quantity:        3,
		PriceAtPurchase: 99.99,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price_at_purchase", "created_at"}).
		AddRow(1, item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO order_items`).
		WithArgs(item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase).
		WillReturnRows(rows)
	mock.ExpectCommit()

	result, err := repo.CreateOrderItem(context.Background(), item)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, item.PriceAtPurchase, result.PriceAtPurchase)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrderItem_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	item := &entity.OrderItem{
		OrderID:         1,
		ProductID:       2,
		Quantity:        3,
		PriceAtPurchase: 99.99,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO order_items`).
		WithArgs(item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	result, err := repo.CreateOrderItem(context.Background(), item)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderItems_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price_at_purchase", "created_at"}).
		AddRow(1, 1, 2, 3, 99.99, now).
		AddRow(2, 1, 3, 1, 49.99, now)

	mock.ExpectQuery(`SELECT id, order_id, product_id, quantity, price_at_purchase, created_at FROM order_items WHERE order_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetOrderItems(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].OrderID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderItems_Empty(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price_at_purchase", "created_at"})

	mock.ExpectQuery(`SELECT id, order_id, product_id, quantity, price_at_purchase, created_at FROM order_items WHERE order_id = \$1`).
		WithArgs(int64(999)).
		WillReturnRows(rows)

	result, err := repo.GetOrderItems(context.Background(), 999)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetOrderItems_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, order_id, product_id, quantity, price_at_purchase, created_at FROM order_items WHERE order_id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetOrderItems(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_AddToCart_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	item := &entity.CartItem{
		UserID:    1,
		ProductID: 2,
		Quantity:  3,
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "created_at", "updated_at"}).
		AddRow(1, item.UserID, item.ProductID, item.Quantity, now, now)

	mock.ExpectQuery(`INSERT INTO cart_items`).
		WithArgs(item.UserID, item.ProductID, item.Quantity).
		WillReturnRows(rows)

	result, err := repo.AddToCart(context.Background(), item)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, item.Quantity, result.Quantity)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_AddToCart_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	item := &entity.CartItem{
		UserID:    1,
		ProductID: 2,
		Quantity:  3,
	}

	mock.ExpectQuery(`INSERT INTO cart_items`).
		WithArgs(item.UserID, item.ProductID, item.Quantity).
		WillReturnError(errors.New("database error"))

	result, err := repo.AddToCart(context.Background(), item)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartByUserID_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "created_at", "updated_at"}).
		AddRow(1, 1, 2, 3, now, now).
		AddRow(2, 1, 3, 1, now, now)

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	result, err := repo.GetCartByUserID(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartByUserID_Empty(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(999)).
		WillReturnRows(rows)

	result, err := repo.GetCartByUserID(context.Background(), 999)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartByUserID_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetCartByUserID(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartItem_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "product_id", "quantity", "created_at", "updated_at"}).
		AddRow(1, 1, 2, 3, now, now)

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1 AND product_id = \$2`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(rows)

	result, err := repo.GetCartItem(context.Background(), 1, 2)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.UserID)
	assert.Equal(t, int64(2), result.ProductID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartItem_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1 AND product_id = \$2`).
		WithArgs(int64(1), int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetCartItem(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrCartItemNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetCartItem_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, user_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE user_id = \$1 AND product_id = \$2`).
		WithArgs(int64(1), int64(2)).
		WillReturnError(errors.New("database error"))

	result, err := repo.GetCartItem(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateCartItem_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE cart_items SET quantity = \$1 WHERE id = \$2`).
		WithArgs(5, int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateCartItem(context.Background(), 1, 5)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateCartItem_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE cart_items SET quantity = \$1 WHERE id = \$2`).
		WithArgs(5, int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.UpdateCartItem(context.Background(), 999, 5)

	assert.Error(t, err)
	assert.Equal(t, ErrCartItemNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateCartItem_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE cart_items SET quantity = \$1 WHERE id = \$2`).
		WithArgs(5, int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.UpdateCartItem(context.Background(), 1, 5)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_DeleteCartItem_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM cart_items WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteCartItem(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_DeleteCartItem_NotFound(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM cart_items WHERE id = \$1`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repo.DeleteCartItem(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, ErrCartItemNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_DeleteCartItem_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM cart_items WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))
	mock.ExpectRollback()

	err := repo.DeleteCartItem(context.Background(), 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_ClearCart_Success(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.ClearCart(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_ClearCart_EmptyCart(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.ClearCart(context.Background(), 1)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_ClearCart_DBError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM cart_items WHERE user_id = \$1`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("database error"))

	err := repo.ClearCart(context.Background(), 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_CreateOrder_CommitError(t *testing.T) {
	repo, mock, cleanup := setupOrderRepositoryTest(t)
	defer cleanup()

	order := &entity.Order{
		UserID:        1,
		Status:        "pending",
		TotalAmount:   199.99,
		PaymentStatus: "pending",
	}

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "status", "total_amount", "payment_status", "payment_id", "created_at", "updated_at"}).
		AddRow(1, order.UserID, order.Status, order.TotalAmount, order.PaymentStatus, nil, now, now)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs(order.UserID, order.Status, order.TotalAmount, order.PaymentStatus).
		WillReturnRows(rows)
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	result, err := repo.CreateOrder(context.Background(), order)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
