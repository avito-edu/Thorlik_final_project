package dto

import "time"

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1"`
}

type OrderItemRequest struct {
	ProductID int64 `json:"product_id" validate:"required"`
	Quantity  int   `json:"quantity" validate:"required,gt=0"`
}

type OrderResponse struct {
	ID            int64               `json:"id"`
	UserID        int64               `json:"user_id"`
	Status        string              `json:"status"`
	TotalAmount   float64             `json:"total_amount"`
	PaymentStatus string              `json:"payment_status"`
	PaymentID     string              `json:"payment_id,omitempty"`
	Items         []OrderItemResponse `json:"items,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

type OrderItemResponse struct {
	ID              int64   `json:"id"`
	ProductID       int64   `json:"product_id"`
	ProductTitle    string  `json:"product_title,omitempty"`
	Quantity        int     `json:"quantity"`
	PriceAtPurchase float64 `json:"price_at_purchase"`
}

type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type AddToCartRequest struct {
	ProductID int64 `json:"product_id" validate:"required"`
	Quantity  int   `json:"quantity" validate:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0"`
}

type CartItemResponse struct {
	ID        int64           `json:"id"`
	ProductID int64           `json:"product_id"`
	Product   ProductResponse `json:"product,omitempty"`
	Quantity  int             `json:"quantity"`
}

type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}

type PaymentRequest struct {
	OrderID       int64  `json:"order_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
}

type PaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
	TransactionID string `json:"transaction_id,omitempty"`
}
