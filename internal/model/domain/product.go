package domain

import "time"

type ProductStatus string

const (
	ProductStatusPending  ProductStatus = "pending"
	ProductStatusApproved ProductStatus = "approved"
	ProductStatusRejected ProductStatus = "rejected"
)

type Product struct {
	ID          int64         `json:"id,omitempty"`
	SellerID    int64         `json:"seller_id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Price       float64       `json:"price"`
	Quantity    int           `json:"quantity"`
	Category    string        `json:"category"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at,omitempty"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty"`
}

func (p *Product) IsAvailable() bool {
	return p.Status == ProductStatusApproved && p.Quantity > 0
}
