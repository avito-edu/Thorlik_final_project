package dto

import "time"

// CreateOrderRequest represents order creation request
// @Description Request body for creating a new order
type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1"`
}

// OrderItemRequest represents an item in order request
// @Description Single item in order creation request
type OrderItemRequest struct {
	ProductID int64 `json:"product_id" validate:"required" example:"1"`
	Quantity  int   `json:"quantity" validate:"required,gt=0" example:"2"`
}

// OrderResponse represents order data in responses
// @Description Order information response
type OrderResponse struct {
	ID            int64               `json:"id" example:"1"`
	UserID        int64               `json:"user_id" example:"5"`
	Status        string              `json:"status" example:"pending" enums:"pending,processing,shipped,delivered,cancelled"`
	TotalAmount   float64             `json:"total_amount" example:"199.99"`
	PaymentStatus string              `json:"payment_status" example:"pending" enums:"pending,completed,failed,refunded"`
	PaymentID     string              `json:"payment_id,omitempty" example:"pay_abc123"`
	Items         []OrderItemResponse `json:"items,omitempty"`
	CreatedAt     time.Time           `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// OrderItemResponse represents an item in order response
// @Description Single item in order response
type OrderItemResponse struct {
	ID              int64   `json:"id" example:"1"`
	ProductID       int64   `json:"product_id" example:"10"`
	ProductTitle    string  `json:"product_title,omitempty" example:"iPhone 15 Pro"`
	Quantity        int     `json:"quantity" example:"2"`
	PriceAtPurchase float64 `json:"price_at_purchase" example:"999.99"`
}

// OrderListResponse represents paginated list of orders
// @Description Paginated order list response
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count" example:"50"`
	Page       int             `json:"page" example:"1"`
	PageSize   int             `json:"page_size" example:"20"`
}

// UpdateOrderStatusRequest represents order status update request
// @Description Admin request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required" example:"shipped" enums:"pending,processing,shipped,delivered,cancelled"`
}

// AddToCartRequest represents add to cart request
// @Description Request body for adding item to cart
type AddToCartRequest struct {
	ProductID int64 `json:"product_id" validate:"required" example:"1"`
	Quantity  int   `json:"quantity" validate:"required,gt=0" example:"1"`
}

// UpdateCartItemRequest represents cart item update request
// @Description Request body for updating cart item quantity
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gt=0" example:"3"`
}

// CartItemResponse represents cart item in response
// @Description Single item in cart response
type CartItemResponse struct {
	ID        int64           `json:"id" example:"1"`
	ProductID int64           `json:"product_id" example:"10"`
	Product   ProductResponse `json:"product,omitempty"`
	Quantity  int             `json:"quantity" example:"2"`
}

// CartResponse represents user's shopping cart
// @Description User's shopping cart response
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items" example:"3"`
	TotalPrice float64            `json:"total_price" example:"299.99"`
}

// PaymentRequest represents payment processing request
// @Description Request body for processing payment
type PaymentRequest struct {
	OrderID       int64  `json:"order_id" validate:"required" example:"1"`
	PaymentMethod string `json:"payment_method" validate:"required" example:"card" enums:"card,paypal,bank_transfer"`
}

// PaymentResponse represents payment processing response
// @Description Payment processing result
type PaymentResponse struct {
	PaymentID     string `json:"payment_id" example:"pay_abc123"`
	Status        string `json:"status" example:"completed" enums:"completed,pending,failed"`
	Message       string `json:"message" example:"Payment processed successfully"`
	TransactionID string `json:"transaction_id,omitempty" example:"txn_xyz789"`
}
