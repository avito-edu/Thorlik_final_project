package model

import (
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/model/entity"
)

func UserEntityToDomain(e *entity.User) *domain.User {
	if e == nil {
		return nil
	}
	return &domain.User{
		ID:        e.ID,
		Email:     e.Email,
		FirstName: e.FirstName,
		LastName:  e.LastName,
		Role:      domain.UserRole(e.Role),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func UserDomainToEntity(d *domain.User, passwordHash string) *entity.User {
	if d == nil {
		return nil
	}
	return &entity.User{
		ID:           d.ID,
		Email:        d.Email,
		PasswordHash: passwordHash,
		FirstName:    d.FirstName,
		LastName:     d.LastName,
		Role:         string(d.Role),
	}
}

func ProductEntityToDomain(e *entity.Product) *domain.Product {
	if e == nil {
		return nil
	}
	return &domain.Product{
		ID:          e.ID,
		SellerID:    e.SellerID,
		Title:       e.Title,
		Description: e.Description,
		Price:       e.Price,
		Quantity:    e.Quantity,
		Category:    e.Category,
		Status:      domain.ProductStatus(e.Status),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func ProductDomainToEntity(d *domain.Product) *entity.Product {
	if d == nil {
		return nil
	}
	return &entity.Product{
		ID:          d.ID,
		SellerID:    d.SellerID,
		Title:       d.Title,
		Description: d.Description,
		Price:       d.Price,
		Quantity:    d.Quantity,
		Category:    d.Category,
		Status:      string(d.Status),
	}
}

func OrderEntityToDomain(e *entity.Order) *domain.Order {
	if e == nil {
		return nil
	}
	return &domain.Order{
		ID:            e.ID,
		UserID:        e.UserID,
		Status:        domain.OrderStatus(e.Status),
		TotalAmount:   e.TotalAmount,
		PaymentStatus: domain.PaymentStatus(e.PaymentStatus),
		PaymentID:     e.PaymentID,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func OrderDomainToEntity(d *domain.Order) *entity.Order {
	if d == nil {
		return nil
	}
	return &entity.Order{
		ID:            d.ID,
		UserID:        d.UserID,
		Status:        string(d.Status),
		TotalAmount:   d.TotalAmount,
		PaymentStatus: string(d.PaymentStatus),
		PaymentID:     d.PaymentID,
	}
}

func OrderItemEntityToDomain(e *entity.OrderItem) *domain.OrderItem {
	if e == nil {
		return nil
	}
	return &domain.OrderItem{
		ID:              e.ID,
		OrderID:         e.OrderID,
		ProductID:       e.ProductID,
		Quantity:        e.Quantity,
		PriceAtPurchase: e.PriceAtPurchase,
		CreatedAt:       e.CreatedAt,
	}
}

func OrderItemDomainToEntity(d *domain.OrderItem) *entity.OrderItem {
	if d == nil {
		return nil
	}
	return &entity.OrderItem{
		ID:              d.ID,
		OrderID:         d.OrderID,
		ProductID:       d.ProductID,
		Quantity:        d.Quantity,
		PriceAtPurchase: d.PriceAtPurchase,
	}
}

func CartItemEntityToDomain(e *entity.CartItem) *domain.CartItem {
	if e == nil {
		return nil
	}
	return &domain.CartItem{
		ID:        e.ID,
		UserID:    e.UserID,
		ProductID: e.ProductID,
		Quantity:  e.Quantity,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func CartItemDomainToEntity(d *domain.CartItem) *entity.CartItem {
	if d == nil {
		return nil
	}
	return &entity.CartItem{
		ID:        d.ID,
		UserID:    d.UserID,
		ProductID: d.ProductID,
		Quantity:  d.Quantity,
	}
}
