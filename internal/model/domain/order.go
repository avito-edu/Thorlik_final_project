package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusSuccess  PaymentStatus = "success"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type Order struct {
	ID            int64         `json:"id,omitempty"`
	UserID        int64         `json:"user_id"`
	Status        OrderStatus   `json:"status"`
	TotalAmount   float64       `json:"total_amount"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	PaymentID     string        `json:"payment_id,omitempty"`
	Items         []OrderItem   `json:"items,omitempty"`
	CreatedAt     time.Time     `json:"created_at,omitempty"`
	UpdatedAt     time.Time     `json:"updated_at,omitempty"`
}

type OrderItem struct {
	ID              int64     `json:"id,omitempty"`
	OrderID         int64     `json:"order_id"`
	ProductID       int64     `json:"product_id"`
	Quantity        int       `json:"quantity"`
	PriceAtPurchase float64   `json:"price_at_purchase"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

type CartItem struct {
	ID        int64     `json:"id,omitempty"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Product   *Product  `json:"product,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
