package models

import (
	"ecommerce-service/graph/model"
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusCompleted  OrderStatus = "COMPLETED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type Order struct {
	Base
	CustomerID uuid.UUID   `gorm:"type:uuid;not null"`
	Customer   User        `gorm:"foreignkey:CustomerID"`
	Items      []OrderItem `gorm:"foreignkey:OrderID"`
	Status     OrderStatus `gorm:"not null;default:'PENDING'"`
	Total      float64     `gorm:"not null"`
}

type OrderItem struct {
	Base
	OrderID   uuid.UUID `gorm:"type:uuid;not null"`
	ProductID uuid.UUID `gorm:"type:uuid;not null"`
	Product   Product
	Quantity  int     `gorm:"not null"`
	UnitPrice float64 `gorm:"not null"`
	SubTotal  float64 `gorm:"not null"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	type OrderAlias Order
	return json.Marshal(&struct {
		OrderAlias
		ID string `json:"id"`
	}{
		OrderAlias: OrderAlias(o),
		ID:         o.ID.String(),
	})
}

func (o Order) ToGraphQL() *model.Order {
	items := make([]*model.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = item.ToGraphQL()
	}

	return &model.Order{
		ID:        o.ID.String(),
		Customer:  o.Customer.ToGraphData(),
		Items:     items,
		Status:    model.OrderStatus(o.Status),
		Total:     o.Total,
		CreatedAt: o.CreatedAt,
	}
}

func (oi OrderItem) ToGraphQL() *model.OrderItem {
	return &model.OrderItem{
		ID:        oi.ID.String(),
		Product:   oi.Product.ToGraphQL(),
		Quantity:  int32(oi.Quantity),
		UnitPrice: oi.UnitPrice,
		SubTotal:  oi.SubTotal,
	}
}
