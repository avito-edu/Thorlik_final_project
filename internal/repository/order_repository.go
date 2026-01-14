package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Thorlik/marketplace/internal/model/entity"
	"github.com/Thorlik/marketplace/internal/pkg/db"
	"go.uber.org/zap"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrCartItemNotFound = errors.New("cart item not found")
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error)
	GetOrderByID(ctx context.Context, id int64) (*entity.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Order, error)
	GetAllOrders(ctx context.Context, limit, offset int) ([]*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status string) error
	UpdatePaymentStatus(ctx context.Context, id int64, status, paymentID string) error

	CreateOrderItem(ctx context.Context, item *entity.OrderItem) (*entity.OrderItem, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]*entity.OrderItem, error)

	AddToCart(ctx context.Context, item *entity.CartItem) (*entity.CartItem, error)
	GetCartByUserID(ctx context.Context, userID int64) ([]*entity.CartItem, error)
	GetCartItem(ctx context.Context, userID, productID int64) (*entity.CartItem, error)
	UpdateCartItem(ctx context.Context, id int64, quantity int) error
	DeleteCartItem(ctx context.Context, id int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type orderRepository struct {
	txManager *db.TransactionManager
	qManager  *db.QueryManager
	logger    *zap.Logger
}

func NewOrderRepository(txManager *db.TransactionManager, qManager *db.QueryManager, logger *zap.Logger) OrderRepository {
	return &orderRepository{
		txManager: txManager,
		qManager:  qManager,
		logger:    logger,
	}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	r.logger.Info("Creating order", zap.Int64("user_id", order.UserID))

	var createdOrder entity.Order
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			INSERT INTO orders (user_id, status, total_amount, payment_status)
			VALUES ($1, $2, $3, $4)
			RETURNING id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at
		`
		row := tx.QueryRow(ctx, query, order.UserID, order.Status, order.TotalAmount, order.PaymentStatus)
		var paymentID sql.NullString
		if err := row.Scan(&createdOrder.ID, &createdOrder.UserID, &createdOrder.Status,
			&createdOrder.TotalAmount, &createdOrder.PaymentStatus, &paymentID,
			&createdOrder.CreatedAt, &createdOrder.UpdatedAt); err != nil {
			r.logger.Error("Failed to create order", zap.Error(err))
			return fmt.Errorf("failed to create order: %w", err)
		}
		if paymentID.Valid {
			createdOrder.PaymentID = paymentID.String
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.Info("Order created successfully", zap.Int64("id", createdOrder.ID))
	return &createdOrder, nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	r.logger.Debug("Getting order by ID", zap.Int64("id", id))

	query := `
		SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.qManager.QueryRow(ctx, query, id)

	var order entity.Order
	var paymentID sql.NullString
	err := row.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
		&order.PaymentStatus, &paymentID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		r.logger.Error("Failed to get order", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if paymentID.Valid {
		order.PaymentID = paymentID.String
	}

	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Order, error) {
	r.logger.Debug("Getting orders by user ID", zap.Int64("user_id", userID))

	query := `
		SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.qManager.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		var paymentID sql.NullString
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.PaymentStatus, &paymentID, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		if paymentID.Valid {
			order.PaymentID = paymentID.String
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context, limit, offset int) ([]*entity.Order, error) {
	r.logger.Debug("Getting all orders")

	query := `
		SELECT id, user_id, status, total_amount, payment_status, payment_id, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.qManager.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		var paymentID sql.NullString
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.PaymentStatus, &paymentID, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		if paymentID.Valid {
			order.PaymentID = paymentID.String
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	r.logger.Info("Updating order status", zap.Int64("id", id), zap.String("status", status))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE orders SET status = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrOrderNotFound
		}

		return nil
	})

	return err
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id int64, status, paymentID string) error {
	r.logger.Info("Updating payment status", zap.Int64("id", id), zap.String("status", status))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE orders SET payment_status = $1, payment_id = $2 WHERE id = $3`
		result, err := tx.Exec(ctx, query, status, paymentID, id)
		if err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrOrderNotFound
		}

		return nil
	})

	return err
}

func (r *orderRepository) CreateOrderItem(ctx context.Context, item *entity.OrderItem) (*entity.OrderItem, error) {
	r.logger.Debug("Creating order item", zap.Int64("order_id", item.OrderID))

	var createdItem entity.OrderItem
	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `
			INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase)
			VALUES ($1, $2, $3, $4)
			RETURNING id, order_id, product_id, quantity, price_at_purchase, created_at
		`
		row := tx.QueryRow(ctx, query, item.OrderID, item.ProductID, item.Quantity, item.PriceAtPurchase)

		if err := row.Scan(&createdItem.ID, &createdItem.OrderID, &createdItem.ProductID,
			&createdItem.Quantity, &createdItem.PriceAtPurchase, &createdItem.CreatedAt); err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &createdItem, nil
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]*entity.OrderItem, error) {
	r.logger.Debug("Getting order items", zap.Int64("order_id", orderID))

	query := `
		SELECT id, order_id, product_id, quantity, price_at_purchase, created_at
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.qManager.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID,
			&item.Quantity, &item.PriceAtPurchase, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (r *orderRepository) AddToCart(ctx context.Context, item *entity.CartItem) (*entity.CartItem, error) {
	r.logger.Debug("Adding to cart", zap.Int64("user_id", item.UserID), zap.Int64("product_id", item.ProductID))

	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + $3
		RETURNING id, user_id, product_id, quantity, created_at, updated_at
	`
	row := r.qManager.QueryRow(ctx, query, item.UserID, item.ProductID, item.Quantity)

	var createdItem entity.CartItem
	if err := row.Scan(&createdItem.ID, &createdItem.UserID, &createdItem.ProductID,
		&createdItem.Quantity, &createdItem.CreatedAt, &createdItem.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to add to cart: %w", err)
	}

	return &createdItem, nil
}

func (r *orderRepository) GetCartByUserID(ctx context.Context, userID int64) ([]*entity.CartItem, error) {
	r.logger.Debug("Getting cart by user ID", zap.Int64("user_id", userID))

	query := `
		SELECT id, user_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.qManager.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.CartItem
	for rows.Next() {
		var item entity.CartItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ProductID,
			&item.Quantity, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (r *orderRepository) GetCartItem(ctx context.Context, userID, productID int64) (*entity.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2
	`
	row := r.qManager.QueryRow(ctx, query, userID, productID)

	var item entity.CartItem
	err := row.Scan(&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	return &item, nil
}

func (r *orderRepository) UpdateCartItem(ctx context.Context, id int64, quantity int) error {
	r.logger.Debug("Updating cart item", zap.Int64("id", id), zap.Int("quantity", quantity))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `UPDATE cart_items SET quantity = $1 WHERE id = $2`
		result, err := tx.Exec(ctx, query, quantity, id)
		if err != nil {
			return fmt.Errorf("failed to update cart item: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrCartItemNotFound
		}

		return nil
	})

	return err
}

func (r *orderRepository) DeleteCartItem(ctx context.Context, id int64) error {
	r.logger.Debug("Deleting cart item", zap.Int64("id", id))

	err := r.txManager.WithTransaction(ctx, func(tx *db.Transaction) error {
		query := `DELETE FROM cart_items WHERE id = $1`
		result, err := tx.Exec(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete cart item: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return ErrCartItemNotFound
		}

		return nil
	})

	return err
}

func (r *orderRepository) ClearCart(ctx context.Context, userID int64) error {
	r.logger.Debug("Clearing cart", zap.Int64("user_id", userID))

	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := r.qManager.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}
