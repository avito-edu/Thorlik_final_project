package entity

import "time"

type Order struct {
	ID            int64     `db:"id"`
	UserID        int64     `db:"user_id"`
	Status        string    `db:"status"`
	TotalAmount   float64   `db:"total_amount"`
	PaymentStatus string    `db:"payment_status"`
	PaymentID     string    `db:"payment_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type OrderItem struct {
	ID              int64     `db:"id"`
	OrderID         int64     `db:"order_id"`
	ProductID       int64     `db:"product_id"`
	Quantity        int       `db:"quantity"`
	PriceAtPurchase float64   `db:"price_at_purchase"`
	CreatedAt       time.Time `db:"created_at"`
}

type CartItem struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	ProductID int64     `db:"product_id"`
	Quantity  int       `db:"quantity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
