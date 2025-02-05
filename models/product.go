package models

import (
	"ecommerce-service/graph/model"
	"encoding/json"
)

type Product struct {
	Base
	Name        string `gorm:"not null"`
	Description string
	Price       float64    `gorm:"not null"`
	SKU         string     `gorm:"uniqueIndex"`
	Categories  []Category `gorm:"many2many:category_products;"`
	Stock       int        `gorm:"not null;default:0"`
}

func (p Product) MarshalJSON() ([]byte, error) {
	type ProductAlias Product
	return json.Marshal(&struct {
		ProductAlias
		ID string `json:"id"`
	}{
		ProductAlias: ProductAlias(p),
		ID:           p.ID.String(),
	})
}

func (p Product) ToGraphQL() *model.Product {
	categories := make([]*model.Category, len(p.Categories))
	for i, cat := range p.Categories {
		categories[i] = &model.Category{
			ID:   cat.ID.String(),
			Name: cat.Name,
		}
	}

	return &model.Product{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: &p.Description,
		Price:       p.Price,
		Sku:         p.SKU,
		Categories:  categories,
		Stock:       int32(p.Stock),
		CreatedAt:   p.CreatedAt,
	}
}
