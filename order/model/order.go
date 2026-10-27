package model

import "time"

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusConfirmed, StatusCancelled:
		return true
	default:
		return false
	}
}

type Order struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	CustomerID  uint        `gorm:"not null;index" json:"customerId" binding:"required"`
	ProductID   uint        `gorm:"not null" json:"productId" binding:"required"`
	Quantity    int         `gorm:"not null" json:"quantity" binding:"required,gt=0"`
	UnitPrice   float64     `gorm:"not null" json:"unitPrice"`
	TotalAmount float64     `gorm:"not null" json:"totalAmount"`
	Status      OrderStatus `gorm:"type:varchar(20);not null;default:'CREATED'" json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CreateOrderInput struct {
	CustomerID uint `json:"customerId" binding:"required"`
	ProductID  uint `json:"productId" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateStatusInput struct {
	Status OrderStatus `json:"status" binding:"required"`
}