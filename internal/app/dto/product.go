package dto

import "time"

type CreateProductRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Quantity    int     `json:"quantity" validate:"required,gte=0"`
	Category    string  `json:"category"`
}

type UpdateProductRequest struct {
	Title       string  `json:"title" validate:"min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gt=0"`
	Quantity    int     `json:"quantity" validate:"gte=0"`
	Category    string  `json:"category"`
}

type ProductResponse struct {
	ID          int64     `json:"id"`
	SellerID    int64     `json:"seller_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type ProductFilter struct {
	Category  string  `json:"category"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	SellerID  int64   `json:"seller_id"`
	Status    string  `json:"status"`
	Page      int     `json:"page"`
	PageSize  int     `json:"page_size"`
	SortBy    string  `json:"sort_by"`
	SortOrder string  `json:"sort_order"`
}

type ModerateProductRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}
