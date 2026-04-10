package entity

import "time"

type Product struct {
	ID          int64     `db:"id"`
	SellerID    int64     `db:"seller_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Quantity    int       `db:"quantity"`
	Category    string    `db:"category"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
