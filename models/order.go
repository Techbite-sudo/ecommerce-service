package models

import uuid "github.com/satori/go.uuid"

// Order represents an order in the system
type Order struct {
	Base
	CustomerID  uuid.UUID   `gorm:"type:uuid;not null"`
	Customer    Customer    `gorm:"foreignkey:CustomerID"`
	Items       []OrderItem `gorm:"foreignkey:OrderID"`
	TotalAmount float64     `gorm:"not null"`
	Status      OrderStatus `gorm:"not null"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	Base
	OrderID   uuid.UUID `gorm:"type:uuid;not null"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	Product   Product   `gorm:"foreignkey:ProductID"`
	Quantity  int       `gorm:"not null"`
	Price     float64   `gorm:"not null"` // Price at time of order
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusDelivered OrderStatus = "DELIVERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)
