package models

import uuid "github.com/satori/go.uuid"

// Category represents a product category with hierarchical structure
type Category struct {
	Base
	Name     string     `gorm:"not null"`
	ParentID *uuid.UUID `gorm:"type:uuid"`
	Parent   *Category  `gorm:"foreignkey:ParentID"`
	Children []Category `gorm:"foreignkey:ParentID"`
	Level    int        `gorm:"not null"` // Track the depth level
	Products []Product  `gorm:"many2many:product_categories;"`
}
