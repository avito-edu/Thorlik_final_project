package dto

import "time"

// CreateProductRequest represents product creation request
// @Description Request body for creating a new product
type CreateProductRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255" example:"iPhone 15 Pro"`
	Description string  `json:"description" example:"Latest Apple smartphone with A17 chip"`
	Price       float64 `json:"price" validate:"required,gt=0" example:"999.99"`
	Quantity    int     `json:"quantity" validate:"required,gte=0" example:"50"`
	Category    string  `json:"category" example:"Electronics"`
}

// UpdateProductRequest represents product update request
// @Description Request body for updating a product
type UpdateProductRequest struct {
	Title       string  `json:"title" validate:"min=1,max=255" example:"iPhone 15 Pro Max"`
	Description string  `json:"description" example:"Updated description"`
	Price       float64 `json:"price" validate:"gt=0" example:"1099.99"`
	Quantity    int     `json:"quantity" validate:"gte=0" example:"100"`
	Category    string  `json:"category" example:"Electronics"`
}

// ProductResponse represents product data in responses
// @Description Product information response
type ProductResponse struct {
	ID          int64     `json:"id" example:"1"`
	SellerID    int64     `json:"seller_id" example:"5"`
	Title       string    `json:"title" example:"iPhone 15 Pro"`
	Description string    `json:"description" example:"Latest Apple smartphone"`
	Price       float64   `json:"price" example:"999.99"`
	Quantity    int       `json:"quantity" example:"50"`
	Category    string    `json:"category" example:"Electronics"`
	Status      string    `json:"status" example:"approved" enums:"pending,approved,rejected"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ProductListResponse represents paginated list of products
// @Description Paginated product list response
type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	TotalCount int64             `json:"total_count" example:"100"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"20"`
}

// ProductFilter represents product filtering options
// @Description Query parameters for filtering products
type ProductFilter struct {
	Category  string  `json:"category" example:"Electronics"`
	MinPrice  float64 `json:"min_price" example:"100"`
	MaxPrice  float64 `json:"max_price" example:"1000"`
	SellerID  int64   `json:"seller_id" example:"5"`
	Status    string  `json:"status" example:"approved"`
	Page      int     `json:"page" example:"1"`
	PageSize  int     `json:"page_size" example:"20"`
	SortBy    string  `json:"sort_by" example:"price"`
	SortOrder string  `json:"sort_order" example:"asc" enums:"asc,desc"`
}

// ModerateProductRequest represents product moderation request
// @Description Moderator request to approve or reject a product
type ModerateProductRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected" example:"approved" enums:"approved,rejected"`
}
